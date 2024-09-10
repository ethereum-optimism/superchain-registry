package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	std "github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/stretchr/testify/require"
)

func testRollupConfig(t *testing.T, chain *ChainConfig) {
	standard := std.Config.Params[chain.Superchain].RollupConfig
	require.Equal(t, chain.AltDA, standard.AltDA, "Standard chains use Ethereum L1 calldata or blobs for data availability (altDA not permitted)")
	assertIntInBounds(t, "Block Time", chain.BlockTime, standard.BlockTime)
	assertIntInBounds(t, "Sequencer Window Size", chain.SequencerWindowSize, standard.SequencerWindowSize)
}

// The values contained in the ChainConfig.Optimism struct used to be hardcoded within
// op-geth, not read from registry config files. This test forces chains that ran an
// old version of add-chain to include these fields, since they are now required downstream
func testOptimismConfig(t *testing.T, chain *ChainConfig) {
	require.NotEmpty(t, chain.Optimism, "Optimism config cannot be nil")
	require.NotEmpty(t, chain.Optimism.EIP1559Elasticity, "EIP1559Elasticity cannot be 0")
	require.NotEmpty(t, chain.Optimism.EIP1559Denominator, "EIP1559Denominator cannot be 0")
	if chain.CanyonTime != nil {
		require.NotEmpty(t, chain.Optimism.EIP1559DenominatorCanyon, "EIP1559DenominatorCanyon cannot be 0 if canyon time is set")
	}
}
