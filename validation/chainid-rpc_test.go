package validation

import (
	"context"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func testChainIDFromRPCofChain(t *testing.T, chain *ChainConfig) {
	isExcluded := map[uint64]bool{
		11155421: true, // sepolia-dev-0/oplabs-devnet-0   No Public RPC declared
		11763072: true, // sepolia-dev-0/base-devnet-0     No Public RPC declared
	}
	if isExcluded[chain.ChainID] {
		t.Skip("chain excluded from chain ID RPC check")
	}
	// Create an ethclient connection to the specified RPC URL
	client, err := ethclient.Dial(chain.PublicRPC)
	require.NoError(t, err, "Failed to connect to the Ethereum client at RPC url %s", chain.PublicRPC)
	defer client.Close()

	// Fetch the chain ID
	chainID, err := Retry(client.NetworkID)(context.Background())

	require.NoError(t, err, "Failed to fetch the chain ID")
	require.Equal(t, chain.ChainID, chainID.Uint64(), "Declared a chainId of %s, but RPC returned ID %s", chain.ChainID, chainID.Uint64())
}
