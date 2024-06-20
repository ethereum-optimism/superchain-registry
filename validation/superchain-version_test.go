package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var isSemverAcceptable = func(desired, actual string) bool {
	return desired == actual
}

func TestSuperchainWideContractVersions(t *testing.T) {
	checkSuperchainTargetSatisfiesSemver := func(t *testing.T, superchain *Superchain) {
		rpcEndpoint := superchain.Config.L1.PublicRPC
		require.NotEmpty(t, rpcEndpoint)

		client, err := ethclient.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		desiredSemver, err := SuperchainSemver[superchain.Superchain].VersionFor("ProtocolVersions")
		require.NoError(t, err)
		checkSemverForContract(t, "ProtocolVersions", superchain.Config.ProtocolVersionsAddr, client, desiredSemver)

		isExcludedFromSuperchainConfigCheck := map[string]bool{
			"Mainnet": true, // no version specified
		}

		if isExcludedFromSuperchainConfigCheck[superchain.Config.Name] {
			t.Skipf("%s excluded from SuperChainConfig version check", superchain.Config.Name)
			return
		}

		desiredSemver, err = SuperchainSemver[superchain.Superchain].VersionFor("SuperchainConfig")
		require.NoError(t, err)
		checkSemverForContract(t, "SuperchainConfig", superchain.Config.SuperchainConfigAddr, client, desiredSemver)
	}

	for superchainName, superchain := range Superchains {
		t.Run(superchainName, func(t *testing.T) { checkSuperchainTargetSatisfiesSemver(t, superchain) })
	}
}

func TestContractVersions(t *testing.T) {
	isExcluded := map[uint64]bool{
		8453:     true, // mainnet/base
		11155421: true, // sepolia-dev-0/oplabs-devnet-0
		11763072: true, // sepolia-dev-0/base-devnet-0
	}

	checkOPChainMatchesATag := func(t *testing.T, chain *ChainConfig) {
		rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC
		require.NotEmpty(t, rpcEndpoint)

		client, err := ethclient.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)
		isFPAC := chain.ChainID == 10 || chain.ChainID == 11155420 || chain.ChainID == 11155421 // TODO don't hardcode this
		versions, err := getContractVersionsFromChain(*Addresses[chain.ChainID], client, isFPAC)
		require.NoError(t, err)
		matches, err := findOPContractTag(versions)
		require.NoError(t, err)
		require.NotEmpty(t, matches)
	}

	for chainID, chain := range OPChains {
		chainID, chain := chainID, chain
		t.Run(perChainTestName(chain), func(t *testing.T) {
			t.Parallel()
			if isExcluded[chainID] {
				t.Skipf("chain %d: EXCLUDED from contract version validation", chainID)
			}
			RunOnStandardAndStandardCandidateChains(t, *chain)
			checkOPChainMatchesATag(t, chain)
		})
	}
}

func getContractVersionsFromChain(list AddressList, client *ethclient.Client, isFPAC bool) (ContractVersions, error) {
	// TODO parallelize this fn
	versions := ContractVersions{}
	var err error

	versions.L1CrossDomainMessenger, err = getVersion(context.Background(), common.Address(list.L1CrossDomainMessengerProxy), client)
	if err != nil {
		return versions, err
	}
	versions.L1ERC721Bridge, err = getVersion(context.Background(), common.Address(list.L1ERC721BridgeProxy), client)
	if err != nil {
		return versions, err
	}

	versions.L1StandardBridge, err = getVersion(context.Background(), common.Address(list.L1StandardBridgeProxy), client)
	if err != nil {
		return versions, err
	}

	if !isFPAC {
		versions.L2OutputOracle, err = getVersion(context.Background(), common.Address(list.L2OutputOracleProxy), client)
		if err != nil {
			return versions, err
		}
	} else {
		versions.AnchorStateRegistry, err = getVersion(context.Background(), common.Address(list.AnchorStateRegistryProxy), client)
		if err != nil {
			return versions, err
		}
		versions.DelayedWETH, err = getVersion(context.Background(), common.Address(list.DelayedWETHProxy), client)
		if err != nil {
			return versions, err
		}
		versions.DisputeGameFactory, err = getVersion(context.Background(), common.Address(list.DisputeGameFactoryProxy), client)
		if err != nil {
			return versions, err
		}
		versions.FaultDisputeGame, err = getVersion(context.Background(), common.Address(list.FaultDisputeGame), client)
		if err != nil {
			return versions, err
		}
		versions.MIPS, err = getVersion(context.Background(), common.Address(list.MIPS), client)
		if err != nil {
			return versions, err
		}
		versions.PermissionedDisputeGame, err = getVersion(context.Background(), common.Address(list.PermissionedDisputeGame), client)
		if err != nil {
			return versions, err
		}
		versions.PreimageOracle, err = getVersion(context.Background(), common.Address(list.PreimageOracle), client)
		if err != nil {
			return versions, err
		}
	}

	versions.OptimismMintableERC20Factory, err = getVersion(context.Background(), common.Address(list.OptimismMintableERC20FactoryProxy), client)
	if err != nil {
		return versions, err
	}

	versions.OptimismPortal, err = getVersion(context.Background(), common.Address(list.OptimismPortalProxy), client)
	if err != nil {
		return versions, err
	}

	versions.SystemConfig, err = getVersion(context.Background(), common.Address(list.SystemConfigProxy), client)
	if err != nil {
		return versions, err
	}

	return versions, nil
}

func checkSemverForContract(t *testing.T, contractName string, contractAddress *Address, client *ethclient.Client, desiredSemver string) {
	actualSemver, err := getVersion(context.Background(), common.Address(*contractAddress), client)
	require.NoError(t, err, "Could not get version for %s", contractName)

	assert.Condition(t, func() bool { return isSemverAcceptable(desiredSemver, actualSemver) },
		"%s.version=%s (UNACCEPTABLE desired version %s)", contractName, actualSemver, desiredSemver)
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
	err = fmt.Errorf("no matching tag found %s", pretty)

	matchesTag := func(standard, candidate ContractVersions) bool {
		// Get the reflection value of the struct
		s := reflect.ValueOf(standard)
		c := reflect.ValueOf(candidate)

		// Iterate over each field of the struct
		for i := 0; i < s.NumField(); i++ {
			field := s.Field(i)
			if field.Kind() == reflect.String &&
				field.String() != "" && // Don't match empty fields
				// We can't check this contract:
				// (until this issue resolves https://github.com/ethereum-optimism/client-pod/issues/699#issuecomment-2150970346)
				s.Type().Field(i).Name != "ProtocolVersions" {
				if field.String() != c.Field(i).String() {
					return false
				}
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
