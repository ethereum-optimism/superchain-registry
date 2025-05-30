package manage

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/log"
)

// GenerateChainArtifacts creates a chain config and genesis file for the chain at index idx in the given state file
// using the given shortName (and optionally, name and superchain identifier).
// It writes these files to the staging directory.
func GenerateChainArtifacts(statePath string, wd string, shortName string, name *string, superchain *string, idx int) error {
	st, err := deployer.ReadOpaqueStateFile(statePath)
	if err != nil {
		return fmt.Errorf("failed to read opaque state file: %w", err)
	}
	l1contractsrelease, err := st.ReadL1ContractsLocator()
	if err != nil {
		return fmt.Errorf("failed to read L1 contracts release: %w", err)
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))
	opd, err := deployer.NewOpDeployer(lgr, l1contractsrelease, deployer.CacheDir)
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
	opaqueGenesis, err := opd.InspectGenesis(statePath, strconv.FormatUint(cfg.ChainID, 10))
	if err != nil {
		return fmt.Errorf("failed to get genesis: %w", err)
	}

	stagingDir := paths.StagingDir(wd)

	output.WriteOK("writing chain config")
	if err := paths.WriteTOMLFile(path.Join(stagingDir, shortName+".toml"), cfg); err != nil {
		return fmt.Errorf("failed to write chain config at index %d: %w", idx, err)
	}

	// TODO: determine if the genesis is deterministic through these conversions
	output.WriteOK("writing genesis")
	genesis, err := opaqueToGenesis(opaqueGenesis)
	if err != nil {
		return fmt.Errorf("failed to convert opaque genesis to core.Genesis: %w", err)
	}

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

// Convert OpaqueMapping to core.Genesis
func opaqueToGenesis(opaque *deployer.OpaqueMap) (*core.Genesis, error) {
	// Step 1: Marshal the OpaqueMapping to JSON
	jsonData, err := json.Marshal(opaque)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OpaqueMapping to JSON: %w", err)
	}

	// Step 2: Unmarshal the JSON data into a core.Genesis struct
	var genesis core.Genesis
	if err := json.Unmarshal(jsonData, &genesis); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to Genesis: %w", err)
	}

	return &genesis, nil
}
