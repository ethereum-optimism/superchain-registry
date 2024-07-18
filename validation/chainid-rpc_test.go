package validation

import (
	"context"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"
)

func testChainIDFromRPC(t *testing.T, chain *ChainConfig) {
	skipIfExcluded(t, chain.ChainID)
	// Create an ethclient connection to the specified RPC URL
	client := clients.L2[chain.ChainID]

	// Fetch the chain ID
	chainID, err := Retry(client.NetworkID)(context.Background())

	require.NoError(t, err, "Failed to fetch the chain ID")
	require.Equal(t, chain.ChainID, chainID.Uint64(), "Declared a chainId of %s, but RPC returned ID %s", chain.ChainID, chainID.Uint64())
}
