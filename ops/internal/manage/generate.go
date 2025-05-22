package manage

import (
	"fmt"
	"path"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/inspect"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
)

// GenerateChainArtifacts creates a chain config and genesis file for the chain at index idx in the given state file
// using the given shortName (and optionally, name and superchain identifier).
// It writes these files to the staging directory.
func GenerateChainArtifacts(st deployer.OpaqueMapping, wd string, shortName string, name *string, superchain *string, idx int) error {
	numChains, err := deployer.GetNumChains(st)
	output.WriteOK("inflating chain config %d of %d", idx, numChains)

	cfg, err := InflateChainConfig(st, idx)
	if err != nil {
		return fmt.Errorf("failed to inflate chain config %d of %d: %w", idx, numChains, err)
	}
	cfg.ShortName = shortName

	if name != nil {
		cfg.Name = *name
	}

	if superchain != nil {
		cfg.Superchain = config.Superchain(*superchain)
	}

	output.WriteOK("reading genesis")
	chainId0, err := deployer.GetChainID(st, 0)
	if err != nil {
		return fmt.Errorf("failed to get chain id %d of %d: %w", idx, numChains, err)
	}

	genesis, _, err := inspect.GenesisAndRollup(st, chainId0)
	if err != nil {
		return fmt.Errorf("failed to get genesis %d of %d: %w", idx, numChains, err)
	}

	stagingDir := paths.StagingDir(wd)

	output.WriteOK("writing chain config")
	if err := paths.WriteTOMLFile(path.Join(stagingDir, cfg.ShortName+".toml"), cfg); err != nil {
		return fmt.Errorf("failed to write chain config %d of %d: %w", idx, numChains, err)
	}

	output.WriteOK("writing genesis")
	if err := WriteGenesis(wd, path.Join(stagingDir, cfg.ShortName+".json.zst"), genesis); err != nil {
		return fmt.Errorf("failed to write genesis %d of %d: %w", idx, numChains, err)
	}
	return nil
}
