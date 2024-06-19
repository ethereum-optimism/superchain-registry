package validation

import (
	"context"
	"fmt"
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
		10:        true, // mainnet/op
		919:       true, // sepolia/mode   L1CrossDomainMessengerProxy.version=1.4.1, https://github.com/ethereum-optimism/security-pod/issues/105
		1740:      true, // sepolia/metal  L1CrossDomainMessengerProxy.version=1.4.1, https://github.com/ethereum-optimism/security-pod/issues/105
		1750:      true, // mainnet/metal  L1CrossDomainMessengerProxy.version=1.4.1, https://github.com/ethereum-optimism/security-pod/issues/105
		8453:      true, // mainnet/base
		8866:      true, // mainnet/superlumio L1CrossDomainMessengerProxy.version=1.4.1, https://github.com/ethereum-optimism/security-pod/issues/105
		34443:     true, // mainnet/mode
		84532:     true, // sepolia/base
		90001:     true, // sepolia/race, due to https://github.com/ethereum-optimism/superchain-registry/issues/147
		7777777:   true, // mainnet/zora
		11155420:  true, // sepolia/op
		11155421:  true, // sepolia-dev-0/oplabs-devnet-0
		11763072:  true, // sepolia-dev-0/base-devnet-0
		999999999: true, // sepolia/zora
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
		t.Run(perChainTestName(chain), func(t *testing.T) {
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
		L1CrossDomainMessenger:       "1.4.0",
		L1ERC721Bridge:               "1.1.1",
		L1StandardBridge:             "1.1.0",
		L2OutputOracle:               "1.3.0",
		OptimismMintableERC20Factory: "1.1.0",
		OptimismPortal:               "1.6.0",
		SystemConfig:                 "1.3.0",
		ProtocolVersions:             "1.0.0",
	}

	got, err := findOPContractTag(shouldMatch)
	require.NoError(t, err)
	want := []standard.Tag{"op-contracts/v1.1.0"}
	require.Equal(t, got, want)

	shouldNotMatch := ContractVersions{
		L1CrossDomainMessenger:       "1.4.0",
		L1ERC721Bridge:               "1.1.1",
		L1StandardBridge:             "1.1.0",
		L2OutputOracle:               "1.3.0",
		OptimismMintableERC20Factory: "1.1.0",
		OptimismPortal:               "1.5.0",
		SystemConfig:                 "1.3.0",
		ProtocolVersions:             "1.0.0",
	}
	got, err = findOPContractTag(shouldNotMatch)
	require.Error(t, err)
	want = []standard.Tag{}
	require.Equal(t, got, want)

}

func findOPContractTag(versions ContractVersions) ([]standard.Tag, error) {
	matchingTags := make([]standard.Tag, 0)
	err := fmt.Errorf("no matching tag found %+v", versions)
	for tag := range standard.Versions {
		if standard.Versions[tag] == versions {
			matchingTags = append(matchingTags, tag)
			err = nil
		}
	}
	return matchingTags, err
}
