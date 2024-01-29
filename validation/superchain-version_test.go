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

// TestContractVersions will check that
//   - for each chain in OPChain
//   - for each declared contract "Foo" : version entry in the corresponding superchain's semver.yaml
//   - the chain has a contract FooProxy deployed at the same version
//
// Actual semvers are
// read from the L1 chain RPC provider for the chain in question.
func TestContractVersions(t *testing.T) {

	isExcluded := map[uint64]bool{
		10:           true,
		291:          true,
		420:          true,
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
		11155420:     true,
		11763071:     true,
		999999999:    true,
		129831238013: true,
	}

	isSemverAcceptable := func(desired, actual string) bool {
		return desired == actual
	}

	checkOPChainSatisfiesSemver := func(chain *ChainConfig) {
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
			require.NoErrorf(t, err, "%s/%s.%s.version= UNSPECIFIED (desired version %s)", chain.Superchain, chain.Name, proxyContractName, desiredSemver)

			actualSemver, err := getVersionWithRetries(context.Background(), common.Address(contractAddress), client)
			require.NoErrorf(t, err, "RPC endpoint %s: %s", rpcEndpoint)

			require.Condition(t, func() bool { return isSemverAcceptable(desiredSemver, actualSemver) },
				"%s/%s.%s.version=%s (UNACCEPTABLE desired version %s)", chain.Superchain, chain.Name, proxyContractName, actualSemver, desiredSemver)

			t.Logf("%s/%s.%s.version=%s (acceptable compared to %s)", chain.Superchain, chain.Name, proxyContractName, actualSemver, desiredSemver)

		}
	}

	for chainID, chain := range OPChains {
		if !isExcluded[chainID] {
			checkOPChainSatisfiesSemver(chain)
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
