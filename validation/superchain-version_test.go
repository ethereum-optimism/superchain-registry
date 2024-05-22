package validation

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var BaseStackContractNames = []string{
	"L1CrossDomainMessenger",
	"L1ERC721Bridge",
	"L1StandardBridge",
	"OptimismMintableERC20Factory",
	"OptimismPortal",
	"SystemConfig",
}

var FaultProofContractNames = []string{
	"AnchorStateRegistry",
	"DelayedWETH",
	"DisputeGameFactory",
	"FaultDisputeGame",
	"MIPS",
	"PermissionedDisputeGame",
	"PreimageOracle",
}

var UnproxiedContractNames = []string{
	"FaultDisputeGame",
	"MIPS",
	"PermissionedDisputeGame",
	"PreimageOracle",
}

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
		919:       true, // sepolia/mode   L1CrossDomainMessengerProxy.version=1.4.1, https://github.com/ethereum-optimism/security-pod/issues/105
		1740:      true, // sepolia/metal  L1CrossDomainMessengerProxy.version=1.4.1, https://github.com/ethereum-optimism/security-pod/issues/105
		1750:      true, // mainnet/metal  L1CrossDomainMessengerProxy.version=1.4.1, https://github.com/ethereum-optimism/security-pod/issues/105
		8453:      true, // mainnet/base
		8866:      true, // mainnet/superlumio L1CrossDomainMessengerProxy.version=1.4.1, https://github.com/ethereum-optimism/security-pod/issues/105
		34443:     true, // mainnet/mode
		84532:     true, // sepolia/base
		90001:     true, // sepolia/race, due to https://github.com/ethereum-optimism/superchain-registry/issues/147
		7777777:   true, // mainnet/zora
		11155421:  true, // sepolia-dev-0/oplabs-devnet-0
		999999999: true, // sepolia/zora
	}

	checkOPChainSatisfiesSemver := func(t *testing.T, chain *ChainConfig) {
		isFaultProofChain := map[uint64]bool{
			11155420: true, // TODO don't hardcode this https://github.com/ethereum-optimism/superchain-registry/issues/219
		}

		rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC

		require.NotEmpty(t, rpcEndpoint)

		client, err := ethclient.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		contractNames := BaseStackContractNames
		if isFaultProofChain[chain.ChainID] {
			// The FaultProof contracts are only present in chains that have activated fault proofs
			contractNames = append(contractNames, FaultProofContractNames...)
		} else {
			// The L2OutputOracle contract is only present in chains that have not activated fault proofs
			contractNames = append(contractNames, "L2OutputOracle")
		}

		for _, contractName := range contractNames {
			desiredSemver, err := SuperchainSemver[chain.Superchain].VersionFor(contractName)
			require.NoError(t, err)
			// ASSUMPTION: we will check the version of the implementation via the declared proxy contract
			var localizedContractName string
			if slices.Contains(UnproxiedContractNames, contractName) {
				localizedContractName = contractName
			} else {
				localizedContractName = contractName + "Proxy"
			}
			contractAddress, err := Addresses[chain.ChainID].AddressFor(localizedContractName)
			require.NoErrorf(t, err, "%s/%s.%s.version= UNSPECIFIED", chain.Superchain, chain.Name, localizedContractName)
			checkSemverForContract(t, localizedContractName, &contractAddress, client, desiredSemver)
		}
	}

	for chainID, chain := range OPChains {
		t.Run(perChainTestName(chain), func(t *testing.T) {
			if isExcluded[chainID] {
				t.Skipf("chain %d: EXCLUDED from contract version validation", chainID)
			}
			SkipCheckIfFrontierChain(t, *chain)
			checkOPChainSatisfiesSemver(t, chain)
		})
	}
}

func checkSemverForContract(t *testing.T, contractName string, contractAddress *Address, client *ethclient.Client, desiredSemver string) {
	actualSemver, err := getVersionWithRetries(context.Background(), common.Address(*contractAddress), client)
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
	version, err := isemver.Version(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", addr, err)
	}

	return version, nil
}

// getVersionWithRetries is a wrapper for getVersion
// which retries up to 10 times with exponential backoff.
func getVersionWithRetries(ctx context.Context, addr common.Address, client *ethclient.Client) (string, error) {
	const maxAttempts = 10
	return retry.Do(ctx, maxAttempts, retry.Exponential(), func() (string, error) {
		return getVersion(context.Background(), addr, client)
	})
}
