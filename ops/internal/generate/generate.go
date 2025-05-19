package generate

import (
	"fmt"
	"path"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/inspect"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
)

// Generate creates a chain config and genesis file for the chain at index idx in the given state file
// using the given shortName (and optionally, name and superchain identifier).
// It writes these files to the staging directory.
func Generate(st state.State, wd string, shortName string, name *string, superchain *string, idx int) error {
	output.WriteOK("inflating chain config %d of %d", idx, len(st.AppliedIntent.Chains))

	cfg, err := manage.InflateChainConfig(&st, idx)
	if err != nil {
		return fmt.Errorf("failed to inflate chain config %d of %d: %w", idx, len(st.AppliedIntent.Chains), err)
	}
	cfg.ShortName = shortName

	if name != nil {
		cfg.Name = *name
	}

	if superchain != nil {
		cfg.Superchain = config.Superchain(*superchain)
	}

	output.WriteOK("reading genesis")
	genesis, _, err := inspect.GenesisAndRollup(&st, st.AppliedIntent.Chains[idx].ID)
	if err != nil {
		return fmt.Errorf("failed to get genesis %d of %d: %w", idx, len(st.AppliedIntent.Chains), err)
	}

	stagingDir := paths.StagingDir(wd)

	output.WriteOK("writing chain config")
	if err := paths.WriteTOMLFile(path.Join(stagingDir, cfg.ShortName+".toml"), cfg); err != nil {
		return fmt.Errorf("failed to write chain config %d of %d: %w", idx, len(st.AppliedIntent.Chains), err)
	}

	output.WriteOK("writing genesis")
	if err := manage.WriteGenesis(wd, path.Join(stagingDir, cfg.ShortName+".json.zst"), genesis); err != nil {
		return fmt.Errorf("failed to write genesis %d of %d: %w", idx, len(st.AppliedIntent.Chains), err)
	}
	return nil
}
