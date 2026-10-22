package manage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
)

// GenerateChainArtifacts creates a chain config and genesis file for the chain at index idx in the given state file
// using the given shortName (and optionally, name and superchain identifier).
// It writes these files to the staging directory.
func GenerateChainArtifacts(
	statePath string,
	wd string,
	shortName string,
	name *string,
	superchain *string,
	idx int,
	opDeployerVersion string,
	opDeployerBinDir string,
	l1ContractsVersion string,
) error {
	st, err := deployer.ReadOpaqueStateFile(statePath)
	if err != nil {
		return fmt.Errorf("failed to read opaque state file: %w", err)
	}

	var picker deployer.BinaryPicker
	if opDeployerVersion == "" {
		// If no op-deployer version is specified, use l1ContractsVersion to determine
		// the appropriate one
		picker, err = deployer.WithReleaseBinary(opDeployerBinDir, l1ContractsVersion)
		if err != nil {
			return fmt.Errorf("failed to autodetect binary: %w", err)
		}
	} else {
		// Otherwise, use the specified op-deployer version. The correct state merger to use will be
		// autodetected based on the provided op-deployer version.
		merger, err := deployer.GetStateMerger(opDeployerVersion)
		if err != nil {
			return fmt.Errorf("failed to get state merger: %w", err)
		}

		binPath := deployer.VersionedBinaryPath(opDeployerBinDir, opDeployerVersion)
		picker = deployer.WithFixedBinary(binPath, merger)
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))
	opd, err := deployer.NewOpDeployer(lgr, picker)
	if err != nil {
		return fmt.Errorf("failed to create op-deployer: %w", err)
	}

	output.WriteOK("inflating chain config at index %d", idx)
	cfg, err := InflateChainConfig(opd, st, statePath, idx, l1ContractsVersion)
	if err != nil {
		return fmt.Errorf("failed to inflate chain config at index %d: %w", idx, err)
	}
	cfg.ShortName = shortName

	if name != nil {
		cfg.Name = *name
	}

	if superchain != nil {
		cfg.Superchain = config.Superchain(*superchain)
		// assume when a superchain is created, the chains created with it will
		// want to inherit standard hardfork times
		cfg.SuperchainTime = new(uint64)
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
	if err := writeCompactGenesis(wd, path.Join(stagingDir, shortName+".json.zst"), genesis); err != nil {
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

// writeCompactGenesis writes the genesis JSON that op-deployer printed to genPath, compacted but
// otherwise as is, so that its fields and their order are kept.
func writeCompactGenesis(rootP, genPath string, genesis json.RawMessage) error {
	compact := new(bytes.Buffer)
	if err := json.Compact(compact, genesis); err != nil {
		return fmt.Errorf("failed to compact genesis: %w", err)
	}
	return WriteGenesis(rootP, genPath, compact.Bytes())
}
