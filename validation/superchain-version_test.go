package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/internal/bindings"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func testContractsMatchATag(t *testing.T, chain *ChainConfig) {
	skipIfExcluded(t, chain.ChainID)

	rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint)

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	versions, err := getContractVersionsFromChain(*Addresses[chain.ChainID], client)
	require.NoError(t, err)
	_, err = findOPContractTag(versions)
	require.NoError(t, err)
}

// getContractVersionsFromChain pulls the appropriate contract versions from chain
// using the supplied client (calling the version() method for each contract). It does this concurrently.
func getContractVersionsFromChain(list AddressList, client *ethclient.Client) (ContractVersions, error) {
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

	getVersionAsync := func(contractAddress Address, results *sync.Map, key string, wg *sync.WaitGroup) {
		r, err := getVersion(context.Background(), common.Address(contractAddress), client)
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
			// If the chain does not store this contractAddress
			// we will continue ("storing" the empty string),
			// so that the rest of the check can
			// still take place. This results in a more useful
			// error shown to the user.
			continue
		}
		wg.Add(1)
		go getVersionAsync(a, results, contractAddress, wg)
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
			// this: we may get the version directly from the implemntation, or via a Proxy,
			// but we store it against the implementation name in either case.
			if s.Type().Field(i).Name == k || s.Type().Field(i).Name+"Proxy" == k {
				reflect.ValueOf(&cv).Elem().Field(i).SetString(v.(string))
			}
		}
		return true
	})

	return cv, nil
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

func TestFindOPContractTag(t *testing.T) {
	shouldMatch := ContractVersions{
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

	got, err := findOPContractTag(shouldMatch)
	require.NoError(t, err)
	want := []standard.Tag{"op-contracts/v1.4.0"}
	require.Equal(t, got, want)

	shouldNotMatch := ContractVersions{
		L1CrossDomainMessenger:       "2.3.0",
		L1ERC721Bridge:               "2.1.0",
		L1StandardBridge:             "2.1.0",
		OptimismMintableERC20Factory: "1.9.0",
		OptimismPortal:               "2.5.0",
		SystemConfig:                 "1.12.0",
		ProtocolVersions:             "1.0.0",
		L2OutputOracle:               "1.0.0",
	}
	got, err = findOPContractTag(shouldNotMatch)
	require.Error(t, err)
	want = []standard.Tag{}
	require.Equal(t, got, want)
}

func findOPContractTag(versions ContractVersions) ([]standard.Tag, error) {
	matchingTags := make([]standard.Tag, 0)
	pretty, err := json.MarshalIndent(versions, "", " ")
	if err != nil {
		return matchingTags, err
	}

	prettyStandard, err := json.MarshalIndent(standard.Versions, "", " ")
	if err != nil {
		return matchingTags, err
	}

	err = fmt.Errorf("contract versions %s do not match any standard op-contracts tag %s", pretty, prettyStandard)

	matchesTag := func(standard, candidate ContractVersions) bool {
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

	for tag := range standard.Versions {
		if matchesTag(standard.Versions[tag], versions) {
			matchingTags = append(matchingTags, tag)
			err = nil
		}
	}
	return matchingTags, err
}
