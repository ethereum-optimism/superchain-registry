package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	std "github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/stretchr/testify/require"
)

func testRollupConfig(t *testing.T, chain *ChainConfig) {
	standard := std.Config.Params[chain.Superchain].RollupConfig
	require.Equal(t, chain.Plasma, standard.Plasma, "Standard chains use Ethereum L1 calldata or blobs for data availability (plasma not permitted)")
	assertIntInBounds(t, "Block Time", chain.BlockTime, standard.BlockTime)
	assertIntInBounds(t, "Sequencer Window Size", chain.SequencerWindowSize, standard.SequencerWindowSize)
}
