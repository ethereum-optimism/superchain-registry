package validation

import (
	"context"
	"math/big"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func testGenesisHashOfChain(t *testing.T, chainID uint64) {
	chainConfig, ok := OPChains[chainID]
	if !ok {
		t.Fatalf("no chain with ID %d found", chainID)
	}

	declaredGenesisHash := chainConfig.Genesis.L2.Hash

	// In this step, we call a function from op-geth which utilises
	// superchain.OPChains, superchain.LoadGenesis, superchain.LoadContractBytecode
	// to reconstruct a core.Genesis and compute its block hash.
	// We can then compare that against the declared genesis block hash.
	// TODO core.LoadOPStackGenesis actually already performs the check we are about to do
	// TODO core.LoadOPStackGenesis could be moved to superchain
	computedGenesis, err := core.LoadOPStackGenesis(chainID)
	require.NoError(t, err)

	computedGenesisHash := computedGenesis.ToBlock().Hash()

	require.Equal(t, common.Hash(declaredGenesisHash), computedGenesisHash, "chain %d: Genesis block hash must match computed value", chainID)
}

// This check should apply to ALL chains in the registry, as it protects downstream code (op-geth)
func TestGenesisHash(t *testing.T) {
	isExcluded := map[uint64]bool{
		10: true, // OP Mainnet, requires override (see https://github.com/ethereum-optimism/op-geth/blob/daade41d463b4ff332c6ed955603e47dcd25528b/core/superchain.go#L83-L94)
	}
	for chainID, chain := range OPChains {
		t.Run(perChainTestName(chain), func(t *testing.T) {
			if isExcluded[chain.ChainID] {
				t.Skipf("chain %d: EXCLUDED from Genesis block hash validation", chainID)
			}
			testGenesisHashOfChain(t, chainID)
		})
	}
}

func TestGenesisHashAgainstRPC(t *testing.T) {
	isExcluded := map[uint64]bool{
		11155421: true, // sepolia-dev-0/oplabs-devnet-0   (no public endpoint)
		11763072: true, // sepolia-dev-0/base-devnet-0     (no public endpoint)
	}

	checkOPChainHashAgainstRPC := func(t *testing.T, chain *ChainConfig) {
		declaredGenesisHash := chain.Genesis.L2.Hash
		rpcEndpoint := chain.PublicRPC
		require.NotEmpty(t, rpcEndpoint)

		client, err := ethclient.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		genesisBlock, err := client.BlockByNumber(context.Background(), big.NewInt(int64(chain.Genesis.L2.Number)))
		require.NoError(t, err)

		require.Equal(t, genesisBlock.Hash(), common.Hash(declaredGenesisHash), "Genesis Block Hash declared as %s, but RPC returned %s", declaredGenesisHash, genesisBlock.Hash())
	}

	for _, chain := range OPChains {
		t.Run(perChainTestName(chain), func(t *testing.T) {
			if isExcluded[chain.ChainID] {
				t.Skip("chain excluded from check")
			}
			checkOPChainHashAgainstRPC(t, chain)
		})
	}
}
