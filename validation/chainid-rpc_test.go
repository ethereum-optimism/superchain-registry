package validation

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-service/retry"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func TestChainIdRPC(t *testing.T) {
	isExcluded := map[uint64]bool{
		11155421: true, // sepolia-dev-0/oplabs-devnet-0   No Public RPC declared
		11763072: true, // sepolia-dev-0/base-devnet-0     No Public RPC declared
	}

	for declaredChainID, chain := range OPChains {
		t.Run(perChainTestName(chain), func(t *testing.T) {
			if isExcluded[declaredChainID] {
				t.Skip("chain excluded from chain ID RPC check")
			}
			// Create an ethclient connection to the specified RPC URL
			client, err := ethclient.Dial(chain.PublicRPC)
			require.NoError(t, err, "Failed to connect to the Ethereum client at RPC url %s", chain.PublicRPC)
			defer client.Close()

			// Fetch the chain ID
			const maxAttempts = 3
			ctx := context.Background()
			chainID, err := retry.Do(ctx, maxAttempts, retry.Exponential(),
				func() (*big.Int, error) {
					return client.NetworkID(ctx)
				})

			require.NoError(t, err, "Failed to fetch the chain ID")
			require.Equal(t, declaredChainID, chainID.Uint64(), "Declared a chainId of %s, but RPC returned ID %s", declaredChainID, chainID.Uint64())
		})
	}
}
