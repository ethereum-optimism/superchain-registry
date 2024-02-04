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

func checkSemverForContract(t *testing.T, contractName string, contractAddress *Address, client *ethclient.Client, desiredSemver string) {

	actualSemver, err := getVersionWithRetries(context.Background(), common.Address(*contractAddress), client)
	require.NoError(t, err, "%s.version= UNSPECIFIED (desired version %s)", contractName, desiredSemver)

	require.Condition(t, func() bool { return isSemverAcceptable(desiredSemver, actualSemver) },
		"%s.version=%s (UNACCEPTABLE desired version %s)", contractName, actualSemver, desiredSemver)

	t.Logf("%s.version=%s (acceptable compared to %s)", contractName, actualSemver, desiredSemver)
}
func TestSuperchainWideContractVersions(t *testing.T) {
	for superchainName, superchain := range Superchains {
		t.Run(superchainName, func(t *testing.T) {
			rpcEndpoint := superchain.Config.L1.PublicRPC
			require.NotEmpty(t, rpcEndpoint)

			client, err := ethclient.Dial(rpcEndpoint)
			require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

			contractNames := []string{
				"ProtocolVersions",
			}

			for _, contractName := range contractNames {
				var contractAddress *Address
				switch contractName {
				case "ProtocolVersions":
					contractAddress = superchain.Config.ProtocolVersionsAddr
				}
				desiredSemver, err := SuperchainSemver[superchainName].VersionFor(contractName)
				require.NoError(t, err)
				checkSemverForContract(t, contractName, contractAddress, client, desiredSemver)
			}
		})
	}
}

func TestContractVersions(t *testing.T) {
	isExcluded := map[uint64]bool{
		291:          true,
		424:          true,
		888:          true,
		957:          true,
		997:          true,
		8453:         true,
		34443:        true,
		58008:        true,
		84531:        true,
		84532:        true,
		7777777:      true,
		11155421:     true, // sepolia-dev-0/oplabs-devnet-0
		11763071:     true,
		999999999:    true,
		129831238013: true,
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
			"L2OutputOrcale",
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
