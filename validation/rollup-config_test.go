package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/stretchr/testify/require"
)

func testRollupConfig(t *testing.T, chain *ChainConfig) {
	require.Equal(t, chain.Plasma, standard.Config.Params[chain.Superchain].RollupConfig.Plasma, "Standard chains use Ethereum L1 calldata or blobs for data availability (plasma not permitted)")
	assertIntInBounds(t, "Block Time", chain.BlockTime, standard.Config.Params[chain.Superchain].RollupConfig.BlockTime)
}
