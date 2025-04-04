package manage

import (
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

	require.Equal(t, []DiskChainConfig{
		{
			ShortName: "op",
			Filepath:  paths.ChainConfig("testdata", "sepolia", "op"),
			Config:    &opConfig,
		},
		{
			ShortName: "testchain",
			Filepath:  paths.ChainConfig("testdata", "sepolia", "testchain"),
			Config:    &testChainConfig,
		},
	}, chains)
}

func TestFindChainConfig(t *testing.T) {
	t.Run("Chain exists", func(t *testing.T) {
		cfg, superchain, err := FindChainConfig("testdata", 11155420)

		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.NotNil(t, cfg.Config.Addresses.SystemConfigProxy)
		require.NotNil(t, cfg.Config.Addresses.L1StandardBridgeProxy)
		require.NotNil(t, cfg.Config.Roles.ProxyAdminOwner)
		require.Equal(t, uint64(11155420), cfg.Config.ChainID)
		require.Equal(t, config.SepoliaSuperchain, superchain)
	})

	t.Run("Chain does not exist", func(t *testing.T) {
		nonExistentChainID := uint64(99999)
		cfg, superchain, err := FindChainConfig("testdata", nonExistentChainID)

		require.Error(t, err)
		require.Nil(t, cfg)
		require.Empty(t, superchain)
		require.Contains(t, err.Error(), "not found")
	})
}
