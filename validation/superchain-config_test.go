package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"
)

func testSuperchainConfig(t *testing.T, chain *ChainConfig) {
	skipIfExcluded(t, chain.ChainID)
	expected := Superchains[chain.Superchain].Config.SuperchainConfigAddr
	require.NotNil(t, expected, "Superchain does not declare a superchain_config_addr")

	client := clients.L1[chain.Superchain]
	defer client.Close()

	opp := Addresses[chain.ChainID].OptimismPortalProxy

	got, err := getAddress("superchainConfig()", opp, client)
	require.NoError(t, err)

	require.Equal(t, *expected, got)
}
