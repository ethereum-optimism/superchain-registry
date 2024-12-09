package validation

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/internal/bindings"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"golang.org/x/mod/semver"
)

func checkForStandardVersions(t *testing.T, chain *ChainConfig) {
	rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint)

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	// testnets and devnets are permitted to use newer contract versions
	// than the versions specified in the standard config
	isTestnet := (chain.Superchain == "sepolia" || chain.Superchain == "sepolia-dev-0")

	versions, err := getContractVersionsFromChain(*Addresses[chain.ChainID], client, chain)
	require.NoError(t, err)
	requireStandardSemvers(t, versions, isTestnet, chain)

	// don't perform bytecode checking for testnets
	if !isTestnet {
		bytecodeHashes, err := getContractBytecodeHashesFromChain(chain.ChainID, *Addresses[chain.ChainID], client, chain)
		require.NoError(t, err)
		requireStandardByteCodeHashes(t, bytecodeHashes, chain)
	}
}

// getContractVersionsFromChain pulls the appropriate contract versions from the chain
// using the supplied client (calling the version() method for each contract). It does this concurrently.
func getContractVersionsFromChain(list AddressList, client *ethclient.Client, chain *ChainConfig) (ContractVersions, error) {
	// Prepare a concurrency-safe object to store version information in, and
	// spin up a goroutine for each contract we are checking (to speed things up).
	results := new(sync.Map)

	getVersionAsync := func(contractAddress Address, results *sync.Map, key string, wg *sync.WaitGroup) {
		r, err := getVersion(context.Background(), common.Address(contractAddress), client)
		if err != nil {
			panic(err)
		}
		results.Store(key, r)
		wg.Done()
	}

	wg := new(sync.WaitGroup)

	contractsToCheckVersionOf := standard.NetworkVersions[chain.Superchain].Releases[standard.Release].GetNonEmpty()

	for _, contractName := range contractsToCheckVersionOf {
		a, err := list.AddressFor(contractName)
		if err != nil {
			// This could be a proxied contract:
			a, err = list.AddressFor(contractName + "Proxy")
			if err != nil {
				panic(fmt.Sprintf("could not find address for %s", contractName))
			}
		}
		wg.Add(1)
		go getVersionAsync(a, results, contractName, wg)
	}

	wg.Wait()

	// use reflection to convert results mapping into a ContractVersions object
	// without resorting to boilerplate code.
	cv := ContractVersions{}
	results.Range(func(k, v any) bool {
		s := reflect.ValueOf(cv)
		for i := 0; i < s.NumField(); i++ {
			// The keys of the results mapping come from the AddressList type,
			// which includes both proxied and unproxied contracts.
			// The cv object (of type ContractVersions), on the other hand,
			// only lists implementation contract versions. The next line accounts for
			// this: we may get the version directly from the implementation, or via a Proxy,
			// but we store it against the implementation name in either case.
			if s.Type().Field(i).Name == k || s.Type().Field(i).Name+"Proxy" == k {
				// there is an inner struct, a VersionedContract - set the `Version`
				reflect.ValueOf(&cv).Elem().Field(i).FieldByName("Version").SetString(v.(string))
			}
		}
		return true
	})

	return cv, nil
}

// getContractBytecodeHashesFromChain pulls the appropriate bytecode from chain
// using the supplied client, concurrently.
func getContractBytecodeHashesFromChain(chainID uint64, list AddressList, client *ethclient.Client, chain *ChainConfig) (standard.L1ContractBytecodeHashes, error) {
	// Prepare a concurrency-safe object to store bytecode information in, and
	// spin up a goroutine for each contract we are checking (to speed things up).
	results := new(sync.Map)

	getBytecodeHashAsync := func(chainID uint64, contractAddress Address, results *sync.Map, contractName string, wg *sync.WaitGroup) {
		r, err := getBytecodeHash(context.Background(), chainID, contractName, common.Address(contractAddress), client)
		if err != nil {
			panic(err)
		}
		results.Store(contractName, r)
		wg.Done()
	}

	wg := new(sync.WaitGroup)

	contractsToCheckBytecodeOf := standard.BytecodeHashes[standard.Release].GetNonEmpty()

	for _, contractName := range contractsToCheckBytecodeOf {
		contractAddress, err := list.AddressFor(contractName)
		if err != nil {
			// This could be a proxied contract:
			contractAddress, err = list.AddressFor(contractName + "Proxy")
			if err != nil {
				panic(fmt.Sprintf("could not find address for %s", contractName))
			}
			contractName = contractName + "Proxy"
		}
		wg.Add(1)
		go getBytecodeHashAsync(chainID, contractAddress, results, contractName, wg)
	}

	wg.Wait()

	// use reflection to convert results mapping into a ContractVersions object
	// without resorting to boilerplate code.
	cbh := standard.L1ContractBytecodeHashes{}
	results.Range(func(k, v any) bool {
		s := reflect.ValueOf(cbh)
		for i := 0; i < s.NumField(); i++ {
			// The keys of the results mapping come from the AddressList type,
			// which includes both proxied and unproxied contracts.
			// The cbh object (of type L1ContractBytecodeHashes), on the other hand,
			// only lists implementation contract bytecode hashes. The next line accounts for
			// this: we may get the bytecode hash directly from the implementation, or via a Proxy,
			if s.Type().Field(i).Name == k || s.Type().Field(i).Name+"Proxy" == k {
				reflect.ValueOf(&cbh).Elem().Field(i).SetString(v.(string))
			}
		}
		return true
	})

	return cbh, nil
}

// getVersion will get the version of a contract at a given address, if it exposes a version() method.
func getVersion(ctx context.Context, addr common.Address, client *ethclient.Client) (string, error) {
	isemver, err := bindings.NewISemver(addr, client)
	if err != nil {
		return "", fmt.Errorf("%s: %w", addr, err)
	}
	version, err := Retry(isemver.Version)(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", addr, err)
	}

	return version, nil
}

// getContractImplAddr gets the implementation contract's address from a deployment of `ProxyAdmin` contract
func getContractImplAddr(
	proxyAdminAddress common.Address,
	targetContractAddr common.Address,
	client *ethclient.Client,
) (common.Address, error) {
	// We need the ABI for ProxyAdmin contract's `getProxyImplementation()`
	// to retrieve the implementation contract's address
	proxyImplABIJson := `[{"inputs":[{"internalType":"address","name":"_proxy","type":"address"}],"name":"getProxyImplementation","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"}]`

	proxyABI, err := abi.JSON(strings.NewReader(proxyImplABIJson))
	if err != nil {
		return common.Address{}, fmt.Errorf("%s: %w", proxyAdminAddress, err)
	}

	var methodData []byte
	if methodData, err = proxyABI.Pack("getProxyImplementation", targetContractAddr); err != nil {
		return common.Address{}, fmt.Errorf("%s: %w", targetContractAddr, err)
	}

	callMsg := ethereum.CallMsg{
		To:   &proxyAdminAddress,
		Data: methodData,
	}

	// Make the call
	callContract := func(msg ethereum.CallMsg) ([]byte, error) {
		return client.CallContract(context.Background(), msg, nil)
	}
	result, err := Retry(callContract)(callMsg)
	if err != nil {
		return common.Address{}, err
	}

	return common.BytesToAddress(result), nil
}

// getBytecodeHash gets the hash of the bytecode of a contract
//   - at a given address, if the contract is not a proxy contract
//   - at the proxy implementation contract's address, if the contract is a proxy contract (we currently use the name suffix to determine
//     whether the contract is a proxy or not)
func getBytecodeHash(ctx context.Context, chainID uint64, contractName string, targetContractAddr common.Address, client *ethclient.Client) (string, error) {
	addrToCheck := targetContractAddr
	proxyContract := strings.HasSuffix(strings.ToLower(contractName), "proxy")
	if proxyContract {
		proxyAdminAddr := Addresses[chainID].ProxyAdmin
		implAddr, err := getContractImplAddr(common.Address(proxyAdminAddr), targetContractAddr, client)
		if err != nil {
			return "", fmt.Errorf("%s/%s: %w", contractName, proxyAdminAddr, err)
		}
		addrToCheck = implAddr
	}

	code, err := client.CodeAt(ctx, addrToCheck, nil)
	if err != nil {
		return "", fmt.Errorf("%s: %w", addrToCheck, err)
	}

	// if the contract is known to have immutables, set up the filterer to mask the bytes which contain the variable's value
	tag := standard.Release
	bytecodeImmutableFilterer, err := initBytecodeImmutableMask(code, tag, contractName)
	// error indicates that the contract _does_ have immutables, but we weren't able to determine the coordinates of the immutables in the bytecode
	if err != nil {
		return "", fmt.Errorf("unable to check for presence of immutables in bytecode: %w", err)
	}

	// For any deployed contracts with immutable variables, the bytecode is masked inside maskBytecode(). If not, the bytecode is unaltered.
	err = bytecodeImmutableFilterer.maskBytecode(contractName)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve bytecode without immutables: %w", err)
	}

	return crypto.Keccak256Hash(bytecodeImmutableFilterer.Bytecode).Hex(), nil
}

func requireStandardSemvers(t *testing.T, versions ContractVersions, isTestnet bool, chain *ChainConfig) {
	standardVersions := standard.NetworkVersions[chain.Superchain].Releases[standard.Release]
	s := reflect.ValueOf(standardVersions)
	c := reflect.ValueOf(versions)
	matches := checkMatchOrTestnet(s, c, isTestnet)

	if !matches {
		diff := cmp.Diff(standardVersions, versions, cmp.FilterPath(func(p cmp.Path) bool {
			return p.Last().String() == ".Address" || p.Last().String() == ".ImplementationAddress"
		}, cmp.Ignore()))
		require.Truef(t, matches,
			"contract versions do not match the standard versions for the %s release \n (-removed from standard / +added to actual):\n %s",
			standard.Release, diff)
	}
}

func requireStandardByteCodeHashes(t *testing.T, hashes standard.L1ContractBytecodeHashes, chain *ChainConfig) {
	standardHashes := standard.BytecodeHashes[standard.Release]
	s := reflect.ValueOf(standardHashes)
	c := reflect.ValueOf(hashes)
	matches := checkMatch(s, c)

	if !matches {
		diff := cmp.Diff(standardHashes, hashes)
		require.Truef(t, matches,
			"contract bytecode hashes do not match the standard bytecode hashes for the %s release \n (-removed from standard / +added to actual):\n %s",
			standard.Release, diff)
	}
}

func checkMatch(s, c reflect.Value) bool {
	// Iterate over each field of the standard struct
	for i := 0; i < s.NumField(); i++ {

		if s.Type().Field(i).Name == "ProtocolVersions" {
			// We can't check this contract:
			// (until this issue resolves https://github.com/ethereum-optimism/client-pod/issues/699#issuecomment-2150970346)
			continue
		}

		field := s.Field(i)

		if field.Kind() != reflect.String {
			panic("versions must be strings")
		}

		if field.String() == "" {
			// Ignore any empty strings, these are treated as "match anything"
			continue
		}

		if field.String() != c.Field(i).String() {
			return false
		}
	}
	return true
}

// checkMatchOrTestnet returns true if s and c match, OR if the chain is a testnet and s < c
func checkMatchOrTestnet(s, c reflect.Value, isTestnet bool) bool {
	// Iterate over each field of the standard struct
	for i := 0; i < s.NumField(); i++ {

		if s.Type().Field(i).Name == "ProtocolVersions" {
			// We can't check this contract:
			// (until this issue resolves https://github.com/ethereum-optimism/client-pod/issues/699#issuecomment-2150970346)
			continue
		}

		innerStStd := s.Field(i).Interface().(VersionedContract)
		innerFieldS := reflect.ValueOf(innerStStd).FieldByName("Version")
		innerStCand := c.Field(i).Interface().(VersionedContract)
		innerFieldC := reflect.ValueOf(innerStCand).FieldByName("Version")

		if innerFieldS.Kind() != reflect.String {
			panic("versions must be strings")
		}

		if innerFieldS.String() == "" {
			// Ignore any empty strings, these are treated as "match anything"
			continue
		}

		if innerFieldS.String() != innerFieldC.String() {
			if !isTestnet {
				return false
			}

			// testnets are permitted to have contract versions that are newer than what's specified in the standard config
			// testnets may NOT have contract versions that are older.
			min := CanonicalizeSemver(innerFieldS.String())
			current := CanonicalizeSemver(innerFieldC.String())
			if semver.Compare(min, current) > 0 {
				return false
			}

		}
	}
	return true
}
