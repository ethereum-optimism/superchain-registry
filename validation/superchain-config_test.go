package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func testSuperchainConfig(t *testing.T, chain *ChainConfig) {
	expected := Superchains[chain.Superchain].Config.SuperchainConfigAddr
	require.NotNil(t, expected, "Superchain does not declare a superchain_config_addr")

	opcm := Superchains[chain.Superchain].Config.OPContractsManagerProxyAddr
	require.NotNil(t, opcm, "Superchain does not declare a op_contracts_manager_proxy_addr")

	rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint, "no rpc specified")

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	checkSuperchainConfig(t, client, Addresses[chain.ChainID].OptimismPortalProxy, *expected)
	checkSuperchainConfig(t, client, Addresses[chain.ChainID].AnchorStateRegistryProxy, *expected)
	checkSuperchainConfig(t, client, Addresses[chain.ChainID].L1CrossDomainMessengerProxy, *expected)
	checkSuperchainConfig(t, client, Addresses[chain.ChainID].L1ERC721BridgeProxy, *expected)
	checkSuperchainConfig(t, client, Addresses[chain.ChainID].L1StandardBridgeProxy, *expected)

	// DelayedWETHProxy uses a different method name, so it is broken out here
	delayedWETHAddress := Addresses[chain.ChainID].DelayedWETHProxy
	got, err := getAddress("config()", delayedWETHAddress, client)
	require.NoError(t, err)
	if *expected != got {
		t.Errorf("incorrect config() address: got %s, wanted %s (queried %s)", got, expected, delayedWETHAddress)
	}
}

func checkSuperchainConfig(t *testing.T, client *ethclient.Client, targetContract Address, expected Address) {
	got, err := getAddress("superchainConfig()", targetContract, client)
	require.NoError(t, err)

	if expected != got {
		t.Errorf("incorrect superchainConfig() address: got %s, wanted %s (queried %s)", got, expected, targetContract)
	}
}
