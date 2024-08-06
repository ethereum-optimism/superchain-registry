package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func testContractsMatchATag(t *testing.T, chain *ChainConfig) {
	skipIfExcluded(t, chain.ChainID)

	rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint)

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	versions, err := getBytecodeHashesFromChain(*Addresses[chain.ChainID], client)
	require.NoError(t, err)
	_, err = findOPContractTag(versions, chain.Superchain)
	require.NoError(t, err)
}

// getBytecodeHashesFromChain pulls the appropriate bytecode from chain
// using the supplied client. It does this concurrently.
func getBytecodeHashesFromChain(list AddressList, client *ethclient.Client) (L1ContractBytecodeHashes, error) {
	// build up list of contracts to check
	contractsToCheck := []string{
		"L1CrossDomainMessengerProxy",
		"L1ERC721BridgeProxy",
		"L1StandardBridgeProxy",
		"OptimismMintableERC20FactoryProxy",
		"OptimismPortalProxy",
		"SystemConfigProxy",
		"AnchorStateRegistryProxy",
		"DelayedWETHProxy",
		"DisputeGameFactoryProxy",
		"FaultDisputeGame",
		"MIPS",
		"PermissionedDisputeGame",
		"PreimageOracle",
	}

	// Prepare a concurrency-safe object to store version information in, and
	// spin up a goroutine for each contract we are checking (to speed things up).
	results := new(sync.Map)

	getBytecodeHashAsync := func(contractAddress Address, results *sync.Map, key string, wg *sync.WaitGroup) {
		r, err := getBytecodeHash(context.Background(), common.Address(contractAddress), client)
		if err != nil {
			panic(err)
		}
		results.Store(key, r)
		wg.Done()
	}

	wg := new(sync.WaitGroup)

	for _, contractAddress := range contractsToCheck {
		a, err := list.AddressFor(contractAddress)
		if err != nil {
			return L1ContractBytecodeHashes{}, err
		}
		wg.Add(1)
		go getBytecodeHashAsync(a, results, contractAddress, wg)
	}

	wg.Wait()

	// use reflection to convert results mapping into a L1ContractBytecodeHashes object
	// without resorting to boilerplate code.
	cbh := L1ContractBytecodeHashes{}
	results.Range(func(k, v any) bool {
		s := reflect.ValueOf(cbh)
		for i := 0; i < s.NumField(); i++ {
			// The keys of the results mapping come from the AddressList type,
			// which includes both proxied and unproxied contracts.
			// The cbh object (of type L1ContractBytecodeHashes), on the other hand,
			// only lists implementation contract bytecode hashes. The next line accounts for
			// this: we may get the bytecode hash directly from the implementation, or via a Proxy,
			// but we store it against the implementation name in either case.
			if s.Type().Field(i).Name == k || s.Type().Field(i).Name+"Proxy" == k {
				reflect.ValueOf(&cbh).Elem().Field(i).SetString(v.(string))
			}
		}
		return true
	})

	return cbh, nil
}

// getBytecodeHash will get the hash of the bytecode of a contract at a given address.
func getBytecodeHash(ctx context.Context, addr common.Address, client *ethclient.Client) (string, error) {
	code, err := client.CodeAt(ctx, addr, nil)

	if err != nil {
		return "", fmt.Errorf("%s: %w", addr, err)
	}

	return crypto.Keccak256Hash(code).Hex(), nil

}

func TestFindOPContractTag(t *testing.T) {
	shouldMatch := L1ContractBytecodeHashes{
		L1CrossDomainMessenger:       "2.3.0",
		L1ERC721Bridge:               "2.1.0",
		L1StandardBridge:             "2.1.0",
		L2OutputOracle:               "",
		OptimismMintableERC20Factory: "1.9.0",
		OptimismPortal:               "3.10.0",
		SystemConfig:                 "2.2.0",
		ProtocolVersions:             "",
		SuperchainConfig:             "",
		AnchorStateRegistry:          "1.0.0",
		DelayedWETH:                  "1.0.0",
		DisputeGameFactory:           "1.0.0",
		FaultDisputeGame:             "1.2.0",
		MIPS:                         "1.0.1",
		PermissionedDisputeGame:      "1.2.0",
		PreimageOracle:               "1.0.0",
	}

	got, err := findOPContractTag(shouldMatch, "mainnet")
	require.NoError(t, err)
	want := []standard.Tag{"op-contracts/v1.4.0"}
	require.Equal(t, got, want)

	shouldNotMatch := L1ContractBytecodeHashes{
		L1CrossDomainMessenger:       "0xaa",
		L1ERC721Bridge:               "0xbb",
		L1StandardBridge:             "0xcc",
		OptimismMintableERC20Factory: "0xdd",
		OptimismPortal:               "0xee",
		SystemConfig:                 "0x00",
		ProtocolVersions:             "0xef",
		L2OutputOracle:               "0x12",
	}
	got, err = findOPContractTag(shouldNotMatch, "mainnet")
	require.Error(t, err)
	want = []standard.Tag{}
	require.Equal(t, got, want)
}

func findOPContractTag(bytecodeHashes L1ContractBytecodeHashes, network string) ([]standard.Tag, error) {
	matchingTags := make([]standard.Tag, 0)
	pretty, err := json.MarshalIndent(bytecodeHashes, "", " ")
	if err != nil {
		return matchingTags, err
	}
	err = fmt.Errorf("no matching tag found %s", pretty)

	matchesTag := func(standard, candidate L1ContractBytecodeHashes) bool {
		s := reflect.ValueOf(standard)
		c := reflect.ValueOf(candidate)

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

	for tag := range *standard.Versions[network] {
		if matchesTag((*standard.Versions[network])[tag], bytecodeHashes) {
			matchingTags = append(matchingTags, tag)
			err = nil
		}
	}
	return matchingTags, err
}
