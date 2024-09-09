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

	rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint, "no rpc specified")

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)
	opp := Addresses[chain.ChainID].OptimismPortalProxy

	got, err := getAddress("superchainConfig()", opp, client)
	require.NoError(t, err)

	if *expected != got {
		t.Errorf("incorrect OptimismPortal.superchainConfig() address: got %s, wanted %s", got, *expected)
	}
}
