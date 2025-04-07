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

func TestFindChainConfigs(t *testing.T) {
	t.Run("Single chain", func(t *testing.T) {
		cfgs, err := FindChainConfigs("testdata", []uint64{11155420})

		require.NoError(t, err)
		require.NotEmpty(t, cfgs)
		require.Equal(t, 1, len(cfgs))
		require.NotNil(t, cfgs[0].Chain.Config.Addresses.SystemConfigProxy)
		require.NotNil(t, cfgs[0].Chain.Config.Addresses.L1StandardBridgeProxy)
		require.NotNil(t, cfgs[0].Chain.Config.Roles.ProxyAdminOwner)
		require.Equal(t, uint64(11155420), cfgs[0].Chain.Config.ChainID)
		require.Equal(t, config.SepoliaSuperchain, cfgs[0].Superchain)
	})

	t.Run("Multiple chains", func(t *testing.T) {
		cfgs, err := FindChainConfigs("testdata", []uint64{11155420, 1952805748})

		require.NoError(t, err)
		require.NotEmpty(t, cfgs)
		require.Equal(t, 2, len(cfgs))
	})

	t.Run("Chain does not exist", func(t *testing.T) {
		nonExistentChainID := uint64(99999)
		_, err := FindChainConfigs("testdata", []uint64{nonExistentChainID})

		require.Error(t, err)
		require.Contains(t, err.Error(), "did not find")
	})
}
