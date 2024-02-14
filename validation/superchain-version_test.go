package validation

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/rpc"

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

		client, err := rpc.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		desiredSemver, err := SuperchainSemver[superchain.Superchain].VersionFor("ProtocolVersions")
		require.NoError(t, err)
		checkSemverForContract(t, "ProtocolVersions", superchain.Config.ProtocolVersionsAddr, client, desiredSemver)
	}

	for superchainName, superchain := range Superchains {
		t.Run(superchainName, func(t *testing.T) { checkSuperchainTargetSatisfiesSemver(t, superchain) })
	}

}

func TestContractVersions(t *testing.T) {

	isExcluded := map[uint64]bool{
		291:          true, // mainnet/orderly
		424:          true, // mainnet/pgn
		888:          true, // goerli-dev-0/op-labs-chaosnet-0 (SystemConfigProxy address not specified)
		957:          true, // mainnet/lyra
		997:          true, // goerli-dev-0/op-labs-devnet-0 (SystemConfigProxy address not specified)
		8453:         true, // mainnet/base
		34443:        true, // mainnet/mode
		58008:        true, // sepolia/pgn
		84531:        true, // goerli/base
		84532:        true, // sepolia/base
		7777777:      true, // mainnet/zora
		11155421:     true, // sepolia-dev-0/oplabs-devnet-0
		11763071:     true, // goerli-dev-0/base-devnet-0 (SystemConfigProxy address not specified)
		999999999:    true, // sepolia/zoras
		129831238013: true, // goerli-dev-0/conduit-devnet-0 (SystemConfigProxy address not specified)
	}

	checkOPChainSatisfiesSemver := func(t *testing.T, chain *ChainConfig) {
		rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC

		require.NotEmpty(t, rpcEndpoint)

		client, err := rpc.Dial(rpcEndpoint)
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

			// ASSUMPTION: we will check the version of the implementation via the declared proxy contract
			proxyContractName := contractName + "Proxy"
			proxyContractAddress, err := Addresses[chain.ChainID].AddressFor(proxyContractName)
			require.NoErrorf(t, err, "%s/%s.%s.version= UNSPECIFIED", chain.Superchain, chain.Name, proxyContractName)

			desiredSemver, err := SuperchainSemver[chain.Superchain].VersionFor(contractName)
			require.NoError(t, err)
			checkSemverForContract(t, proxyContractName, &proxyContractAddress, client, desiredSemver)

			desiredBytecode, ok := ExpectedBytecode[chain.Superchain][contractName]
			if !ok {
				t.Fatalf("no bytecode for %s", contractName)
			}

			checkBytecodeForProxiedContract(t, chain, contractName, &proxyContractAddress, client, desiredBytecode)
		}
	}

	for chainID, chain := range OPChains {
		if isExcluded[chainID] {
			t.Logf("chain %d: EXCLUDED from contract version validation", chainID)
		} else {
			t.Run(chain.Superchain+"/"+chain.Name, func(t *testing.T) { checkOPChainSatisfiesSemver(t, chain) })
		}
	}
}

func checkSemverForContract(t *testing.T, contractName string, contractAddress *Address, client *rpc.Client, desiredSemver string) {
	ethClient := ethclient.NewClient(client)
	actualSemver, err := getVersionWithRetries(context.Background(), common.Address(*contractAddress), ethClient)
	require.NoError(t, err, "Could not get version for %s", contractName)

	require.Condition(t, func() bool { return isSemverAcceptable(desiredSemver, actualSemver) },
		"%s.version=%s (UNACCEPTABLE desired version %s)", contractName, actualSemver, desiredSemver)

	t.Logf("%s.version=%s (acceptable compared to %s)", contractName, actualSemver, desiredSemver)
}

func checkBytecodeForProxiedContract(t *testing.T, chain *ChainConfig, contractName string, contractAddress *Address, client *rpc.Client, desiredBytecode string) {
	actualBytecode, err := getBytecodeForProxiedContract(context.Background(), chain, common.Address(*contractAddress), client)
	require.NoError(t, err, "Could not get bytecode for %s", contractName)

	if diff := cmp.Diff(desiredBytecode, common.Bytes2Hex(actualBytecode)); diff != "" {
		t.Fatalf("bytecode mismatch for %s (-want +got):\n%s", contractName, diff)

	}

	t.Logf("acceptable bytecode for %s", contractName)
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

// getBytecodeWithRetries will get the bytecode for the implementation behind a proxy at the given address, retrying up to 5 times.
func getBytecodeForProxiedContract(ctx context.Context, chain *ChainConfig, proxyAddr common.Address, client *rpc.Client) ([]byte, error) {
	const maxAttempts = 5

	implementationAddr, err := getImplementationAddressForProxy(ctx, chain, proxyAddr, client)
	if err != nil {
		return []byte{}, fmt.Errorf("%s: %w", proxyAddr, err)
	}

	ethClient := ethclient.NewClient(client)
	return retry.Do(ctx, maxAttempts, retry.Exponential(), func() ([]byte, error) {
		return ethClient.CodeAt(ctx, implementationAddr, nil)
	})
}

func getImplementationAddressForProxy(ctx context.Context, chain *ChainConfig, proxyAddr common.Address, client *rpc.Client) (common.Address, error) {
	proxyAdminAddr, err := Addresses[chain.ChainID].AddressFor("ProxyAdmin")
	if err != nil {
		return common.Address{}, err
	}
	proxyAdmin, err := bindings.NewProxyAdmin(common.Address(proxyAdminAddr), ethclient.NewClient(client))
	if err != nil {
		return common.Address{}, err
	}
	maxAttempts := 3
	return retry.Do(ctx, maxAttempts, retry.Exponential(), func() (common.Address, error) {
		return proxyAdmin.GetProxyImplementation(&bind.CallOpts{
			Context: ctx,
		}, proxyAddr)
	},
	)

}
