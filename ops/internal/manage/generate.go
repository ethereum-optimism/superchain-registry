package manage

import (
	"fmt"
	"os"
	"path"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/log"
	"github.com/tomwright/dasel"
)

// GenerateChainArtifacts creates a chain config and genesis file for the chain at index idx in the given state file
// using the given shortName (and optionally, name and superchain identifier).
// It writes these files to the staging directory.
func GenerateChainArtifacts(statePath string, wd string, shortName string, name *string, superchain *string, idx int) error {

	om, err := deployer.ReadOpaqueMappingFile(statePath)
	if err != nil {
		return fmt.Errorf("failed to read opaque mapping file: %w", err)
	}
	l1contractsrelease, err := deployer.ReadL1ContractsRelease(dasel.New(om))
	if err != nil {
		return fmt.Errorf("failed to read L1 contracts release: %w", err)
	}
	chainId, err := deployer.ReadL2ChainId(dasel.New(om), idx)
	if err != nil {
		return fmt.Errorf("failed to read chain ID: %w", err)
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))
	opd, err := deployer.NewOpDeployer(lgr, l1contractsrelease, statePath, wd)
	if err != nil {
		return fmt.Errorf("failed to create op-deployer: %w", err)
	}

	_, err = opd.BuildBinary()
	if err != nil {
		return fmt.Errorf("failed to build op-deployer: %w", err)
	}

	err = opd.SetupOutputState(wd)
	if err != nil {
		return fmt.Errorf("failed to setup output state: %w", err)
	}

	// output.WriteOK("inflating chain config %d of %d", idx, len(st.AppliedIntent.Chains))

	// cfg, err := InflateChainConfig(&st, idx)
	// if err != nil {
	// 	return fmt.Errorf("failed to inflate chain config %d of %d: %w", idx, len(st.AppliedIntent.Chains), err)
	// }
	// cfg.ShortName = shortName

	// if name != nil {
	// 	cfg.Name = *name
	// }

	// if superchain != nil {
	// 	cfg.Superchain = config.Superchain(*superchain)
	// }

	output.WriteOK("reading genesis")

	genesis, err := opd.InspectGenesis(wd, chainId)

	if err != nil {
		return fmt.Errorf("failed to get genesis: %w", err)
	}

	stagingDir := paths.StagingDir(wd)

	// output.WriteOK("writing chain config")
	// if err := paths.WriteTOMLFile(path.Join(stagingDir, cfg.ShortName+".toml"), cfg); err != nil {
	// 	return fmt.Errorf("failed to write chain config %d of %d: %w", idx, len(st.AppliedIntent.Chains), err)
	// }

	output.WriteOK("writing genesis")
	if err := WriteGenesis(wd, path.Join(stagingDir, "test.json.zst"), genesis); err != nil {
		return fmt.Errorf("failed to write genesis: %w", err)
	}
	return nil
}
