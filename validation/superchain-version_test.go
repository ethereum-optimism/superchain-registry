package validation

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/retry"
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
			t.Logf("%s excluded from SuperChainConfig version check", superchain.Config.Name)
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
		291:       true, // mainnet/orderly
		424:       true, // mainnet/pgn
		957:       true, // mainnet/lyra
		8453:      true, // mainnet/base
		34443:     true, // mainnet/mode
		58008:     true, // sepolia/pgn
		84532:     true, // sepolia/base
		7777777:   true, // mainnet/zora
		11155421:  true, // sepolia-dev-0/oplabs-devnet-0
		999999999: true, // sepolia/zoras
	}

	checkOPChainSatisfiesSemver := func(t *testing.T, chain *ChainConfig) {
		rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC

		require.NotEmpty(t, rpcEndpoint)

		client, err := ethclient.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		contractNames := []string{
			"L1CrossDomainMessenger",
			"L1ERC721Bridge",
			"L1StandardBridge",
			"L2OutputOracle",
			"OptimismMintableERC20Factory",
			"OptimismPortal",
			"SystemConfig",
		}

		for _, contractName := range contractNames {
			desiredSemver, err := SuperchainSemver[chain.Superchain].VersionFor(contractName)
			require.NoError(t, err)
			// ASSUMPTION: we will check the version of the implementation via the declared proxy contract
			proxyContractName := contractName + "Proxy"
			contractAddress, err := Addresses[chain.ChainID].AddressFor(proxyContractName)
			require.NoErrorf(t, err, "%s/%s.%s.version= UNSPECIFIED", chain.Superchain, chain.Name, proxyContractName)
			checkSemverForContract(t, proxyContractName, &contractAddress, client, desiredSemver)

		}
	}

	for chainID, chain := range OPChains {
		if isExcluded[chainID] {
			t.Logf("chain %d: EXCLUDED from contract version validation", chainID)
		} else {
			t.Run(chain.Name, func(t *testing.T) { checkOPChainSatisfiesSemver(t, chain) })
		}
	}
}

func checkSemverForContract(t *testing.T, contractName string, contractAddress *Address, client *ethclient.Client, desiredSemver string) {
	actualSemver, err := getVersionWithRetries(context.Background(), common.Address(*contractAddress), client)
	require.NoError(t, err, "Could not get version for %s", contractName)

	require.Condition(t, func() bool { return isSemverAcceptable(desiredSemver, actualSemver) },
		"%s.version=%s (UNACCEPTABLE desired version %s)", contractName, actualSemver, desiredSemver)

	t.Logf("%s.version=%s (acceptable compared to %s)", contractName, actualSemver, desiredSemver)
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
