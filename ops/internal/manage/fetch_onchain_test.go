package manage

import (
	"path/filepath"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/stretchr/testify/require"
)

func TestCollectChainsBySuperchain(t *testing.T) {
	testdataDir, err := filepath.Abs("testdata")
	require.NoError(t, err, "Failed to get absolute path to testdata")

	// Known testdata chainIds
	opSepolia := uint64(11155420)
	testChain := uint64(1952805748)

	t.Run("all chains", func(t *testing.T) {
		chains, err := collectChainsBySuperchain(testdataDir, []uint64{})
		require.NoError(t, err)

		require.Equal(t, len(chains[config.SepoliaSuperchain]), 2)
		require.Equal(t, len(chains[config.MainnetSuperchain]), 0)
		require.Equal(t, len(chains[config.SepoliaDev0Superchain]), 0)
	})

	t.Run("single chain", func(t *testing.T) {
		chains, err := collectChainsBySuperchain(testdataDir, []uint64{opSepolia})
		require.NoError(t, err)

		require.Equal(t, len(chains[config.SepoliaSuperchain]), 1)
		require.Equal(t, len(chains[config.MainnetSuperchain]), 0)
		require.Equal(t, len(chains[config.SepoliaDev0Superchain]), 0)
	})

	t.Run("two chains", func(t *testing.T) {
		chains, err := collectChainsBySuperchain(testdataDir, []uint64{opSepolia, testChain})
		require.NoError(t, err)

		require.Equal(t, len(chains[config.SepoliaSuperchain]), 2)
		require.Equal(t, len(chains[config.MainnetSuperchain]), 0)
		require.Equal(t, len(chains[config.SepoliaDev0Superchain]), 0)
	})

	t.Run("non existent chain", func(t *testing.T) {
		_, err := collectChainsBySuperchain(testdataDir, []uint64{999999999})
		require.Error(t, err)
	})
}
