package manage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, p, content string) {
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
}

// Each reader of registry TOML must reject keys the config structs don't model.
func TestConfigReadersRejectUnknownKeys(t *testing.T) {
	const chainCfg = "name = \"Test\"\nchain_id = 1\n[genesis.system_config]\n  gasLimit = 1\n  not_a_field = 5\n"
	const superchainCfg = "name = \"test\"\nnot_a_field = 5\n"

	tests := []struct {
		name string
		read func(t *testing.T, root string) error
	}{
		{"staged chain config", func(t *testing.T, root string) error {
			writeFile(t, filepath.Join(paths.StagingDir(root), "test.toml"), chainCfg)
			_, err := StagedChainConfigs(root)
			return err
		}},
		{"staged superchain definition", func(t *testing.T, root string) error {
			writeFile(t, filepath.Join(paths.StagingDir(root), "superchain.toml"), superchainCfg)
			_, err := StagedSuperchainDefinition(root)
			return err
		}},
		{"collected chain configs", func(t *testing.T, root string) error {
			writeFile(t, paths.ChainConfig(root, "sepolia", "test"), chainCfg)
			_, err := CollectChainConfigs(paths.SuperchainDir(root, "sepolia"))
			return err
		}},
		{"chain config by name", func(t *testing.T, root string) error {
			writeFile(t, paths.ChainConfig(root, "sepolia", "test"), chainCfg)
			_, err := ReadChainConfig(root, "sepolia", "test")
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.read(t, t.TempDir())
			require.ErrorContains(t, err, "not_a_field")
		})
	}
}

// Every shipped config must decode strictly, so a key added to the repo without its struct field
// fails here rather than in apply_hardforks or a deployer run.
func TestShippedConfigsDecodeStrictly(t *testing.T) {
	wd, err := paths.FindRepoRoot()
	require.NoError(t, err)
	superchains, err := paths.Superchains(wd)
	require.NoError(t, err)
	require.NotEmpty(t, superchains)

	for _, sc := range superchains {
		t.Run(sc, func(t *testing.T) {
			var def config.SuperchainDefinition
			require.NoError(t, paths.ReadTOMLFile(paths.SuperchainConfig(wd, sc), &def))
			chains, err := CollectChainConfigs(paths.SuperchainDir(wd, sc))
			require.NoError(t, err)
			require.NotEmpty(t, chains)
		})
	}
}
