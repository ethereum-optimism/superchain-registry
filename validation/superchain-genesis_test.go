package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
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
	t.Logf("chain %d: Genesis block hash passed validation", chainID)
}

func TestGenesisHash(t *testing.T) {
	isExcluded := map[uint64]bool{
		10: true, // OP Mainnet, requires override (see https://github.com/ethereum-optimism/op-geth/blob/daade41d463b4ff332c6ed955603e47dcd25528b/core/superchain.go#L83-L94)
	}

	for chainID, chain := range OPChains {
		if isExcluded[chain.ChainID] {
			t.Logf("chain %d: EXCLUDED from Genesis block hash validation", chainID)
		} else {
			t.Run(chain.Name, func(t *testing.T) { testGenesisHashOfChain(t, chainID) })
		}
	}
}
