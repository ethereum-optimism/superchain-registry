package manage

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/log"
)

// GenerateChainArtifacts creates a chain config and genesis file for the chain at index idx in the given state file
// using the given shortName (and optionally, name and superchain identifier).
// It writes these files to the staging directory.
func GenerateChainArtifacts(statePath string, wd string, shortName string, name *string, superchain *string, idx int) error {
	st, err := deployer.ReadOpaqueMappingFile(statePath)
	if err != nil {
		return fmt.Errorf("failed to read opaque mapping file: %w", err)
	}
	l1contractsrelease, err := st.ReadL1ContractsLocator()
	if err != nil {
		return fmt.Errorf("failed to read L1 contracts release: %w", err)
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))
	opd, err := deployer.NewOpDeployer(lgr, l1contractsrelease)
	if err != nil {
		return fmt.Errorf("failed to create op-deployer: %w", err)
	}
	output.WriteOK("created op-deployer instance: %s", opd.DeployerVersion)

	output.WriteOK("inflating chain config at index %d", idx)
	cfg, err := InflateChainConfig(opd, st, statePath, idx)
	if err != nil {
		return fmt.Errorf("failed to inflate chain config at index %d: %w", idx, err)
	}
	cfg.ShortName = shortName

	if name != nil {
		cfg.Name = *name
	}

	if superchain != nil {
		cfg.Superchain = config.Superchain(*superchain)
	}

	output.WriteOK("reading genesis")
	genesis, err := opd.InspectGenesis(statePath, strconv.FormatUint(cfg.ChainID, 10))
	if err != nil {
		return fmt.Errorf("failed to get genesis: %w", err)
	}

	stagingDir := paths.StagingDir(wd)

	output.WriteOK("writing chain config")
	if err := paths.WriteTOMLFile(path.Join(stagingDir, shortName+".toml"), cfg); err != nil {
		return fmt.Errorf("failed to write chain config at index %d: %w", idx, err)
	}

	output.WriteOK("writing genesis")
	if err := WriteGenesis(wd, path.Join(stagingDir, shortName+".json.zst"), genesis); err != nil {
		return fmt.Errorf("failed to write genesis at index %d: %w", idx, err)
	}

	// Copy state file to staging directory
	output.WriteOK("writing state.json")
	stateData, err := os.ReadFile(statePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	if err := os.WriteFile(path.Join(stagingDir, "state.json"), stateData, 0o644); err != nil {
		return fmt.Errorf("failed to write state file to staging directory: %w", err)
	}
	return nil
}
