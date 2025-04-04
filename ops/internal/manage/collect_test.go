package manage

import (
	"path/filepath"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/stretchr/testify/require"
)

func TestCollectChainConfigs(t *testing.T) {
	p := paths.SuperchainDir("testdata", "sepolia")

	chains, err := CollectChainConfigs(p)
	require.NoError(t, err)
	require.Len(t, chains, 2)

	var opConfig config.Chain
	var testChainConfig config.Chain
	require.NoError(t, paths.ReadTOMLFile(paths.ChainConfig("testdata", "sepolia", "op"), &opConfig))
	require.NoError(t, paths.ReadTOMLFile(paths.ChainConfig("testdata", "sepolia", "testchain"), &testChainConfig))

	// Create expected chains using the same path generation but normalize the path separators
	expectedChains := []DiskChainConfig{
		{
			ShortName: "op",
			Filepath:  filepath.ToSlash(paths.ChainConfig("testdata", "sepolia", "op")),
			Config:    &opConfig,
		},
		{
			ShortName: "testchain",
			Filepath:  filepath.ToSlash(paths.ChainConfig("testdata", "sepolia", "testchain")),
			Config:    &testChainConfig,
		},
	}

	// Convert actual paths to use forward slashes for consistent comparison
	normalizedChains := make([]DiskChainConfig, len(chains))
	for i, chain := range chains {
		normalizedChains[i] = DiskChainConfig{
			ShortName: chain.ShortName,
			Filepath:  filepath.ToSlash(chain.Filepath),
			Config:    chain.Config,
		}
	}

	require.Equal(t, expectedChains, normalizedChains)
}
