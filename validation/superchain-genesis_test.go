package validation

import (
	"context"
	"math/big"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func testGenesisHash(t *testing.T, chainID uint64) {
	skipIfExcluded(t, chainID)

	chainConfig, ok := OPChains[chainID]
	if !ok {
		t.Fatalf("no chain with ID %d found", chainID)
	}

	declaredGenesisHash := chainConfig.Genesis.L2.Hash

	// In this step, we call a function from op-geth which utilises
	// superchain.OPChains, superchain.LoadGenesis, superchain.LoadContractBytecode
	// to reconstruct a core.Genesis and compute its block hash.
	// We can then compare that against the declared genesis block hash.
	computedGenesis, err := core.LoadOPStackGenesis(chainID)
	require.NoError(t, err)

	computedGenesisHash := computedGenesis.ToBlock().Hash()

	require.Equal(t, common.Hash(declaredGenesisHash), computedGenesisHash, "chain %d: Genesis block hash must match computed value", chainID)
}

func testGenesisHashAgainstRPC(t *testing.T, chain *ChainConfig) {
	skipIfExcluded(t, chain.ChainID)

	declaredGenesisHash := chain.Genesis.L2.Hash

	client := clients.L2[chain.ChainID]
	defer client.Close()

	blockByNumber := func(blockNumber uint64) (*types.Block, error) {
		return client.BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
	}
	genesisBlock, err := Retry(blockByNumber)(chain.Genesis.L2.Number)
	require.NoError(t, err)

	require.Equal(t, genesisBlock.Hash(), common.Hash(declaredGenesisHash), "Genesis Block Hash declared as %s, but RPC returned %s", declaredGenesisHash, genesisBlock.Hash())
}
