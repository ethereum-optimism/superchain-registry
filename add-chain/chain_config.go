package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"gopkg.in/yaml.v3"
)

type JSONChainConfig struct {
	ChainID                          uint64                   `json:"l2_chain_id"`
	BatchInboxAddr                   superchain.Address       `json:"batch_inbox_address"`
	Genesis                          superchain.ChainGenesis  `json:"genesis"`
	PlasmaConfig                     *superchain.PlasmaConfig `json:"plasma_config,omitempty"`
	superchain.HardForkConfiguration `json:",inline"`
}

func (c *JSONChainConfig) VerifyPlasma() error {
	if c.PlasmaConfig != nil {
		if c.PlasmaConfig.DAChallengeAddress == nil {
			return fmt.Errorf("missing required field: da_challenge_contract_address")
		}
		if c.PlasmaConfig.DAChallengeWindow == nil {
			return fmt.Errorf("missing required field: da_challenge_window")
		}
		if c.PlasmaConfig.DAResolveWindow == nil {
			return fmt.Errorf("missing required field: da_resolve_window")
		}
	}
	return nil
}

// constructChainConfig creates and populates a ChainConfig struct by reading from an input file and
// explicitly setting some additional fields to input argument values
func constructChainConfig(
	inputFilePath,
	chainName,
	publicRPC,
	sequencerRPC,
	explorer string,
	superchainLevel superchain.SuperchainLevel,
) (superchain.ChainConfig, error) {
	fmt.Printf("Attempting to read from %s\n", inputFilePath)
	file, err := os.ReadFile(inputFilePath)
	if err != nil {
		return superchain.ChainConfig{}, fmt.Errorf("error reading file: %w", err)
	}
	var jsonConfig JSONChainConfig
	if err = json.Unmarshal(file, &jsonConfig); err != nil {
		return superchain.ChainConfig{}, fmt.Errorf("error unmarshaling json: %w", err)
	}

	err = jsonConfig.VerifyPlasma()
	if err != nil {
		return superchain.ChainConfig{}, fmt.Errorf("error with json plasma config: %w", err)
	}

	chainConfig := superchain.ChainConfig{
		Name:            chainName,
		ChainID:         jsonConfig.ChainID,
		PublicRPC:       publicRPC,
		SequencerRPC:    sequencerRPC,
		Explorer:        explorer,
		BatchInboxAddr:  jsonConfig.BatchInboxAddr,
		Genesis:         jsonConfig.Genesis,
		SuperchainLevel: superchainLevel,
		SuperchainTime:  nil,
		Plasma:          jsonConfig.PlasmaConfig,
		HardForkConfiguration: superchain.HardForkConfiguration{
			CanyonTime:  jsonConfig.CanyonTime,
			DeltaTime:   jsonConfig.DeltaTime,
			EcotoneTime: jsonConfig.EcotoneTime,
			FjordTime:   jsonConfig.FjordTime,
		},
	}

	fmt.Printf("Rollup config successfully constructed\n")
	return chainConfig, nil
}

// writeChainConfig accepts a rollupConfig, formats it, and writes some output files based on the given
// target directories
func writeChainConfig(
	rollupConfig superchain.ChainConfig,
	targetDirectory string,
	superchainRepoPath string,
	superchainTarget string,
) error {
	// Create genesis-system-config data
	// (this is deprecated, users should load this from L1, when available via SystemConfig)
	dirPath := filepath.Join(superchainRepoPath, "superchain", "extra", "genesis-system-configs", superchainTarget)

	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	systemConfigJSON, err := json.MarshalIndent(rollupConfig.Genesis.SystemConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal genesis system config json: %w", err)
	}

	// Write the genesis system config JSON to a new file
	filePath := filepath.Join(dirPath, rollupConfig.Name+".json")
	if err := os.WriteFile(filePath, systemConfigJSON, 0o644); err != nil {
		return fmt.Errorf("failed to write genesis system config json: %w", err)
	}
	fmt.Printf("Genesis system config written to: %s\n", filePath)

	// Remove hardfork timestamp override fields if they match superchain defaults
	defaults := superchain.Superchains[superchainTarget]
	rollupConfig.SetDefaultHardforkTimestampsToNil(&defaults.Config)

	yamlData, err := yaml.Marshal(rollupConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}

	// Unmarshal bytes into a yaml.Node for custom manipulation
	var rootNode yaml.Node
	if err = yaml.Unmarshal(yamlData, &rootNode); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = rollupConfig.EnhanceYAML(ctx, &rootNode); err != nil {
		return err
	}

	// Write the rollup config to a yaml file
	filename := filepath.Join(targetDirectory)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	encoder.SetIndent(2)
	if err := encoder.Encode(&rootNode); err != nil {
		return fmt.Errorf("failed to write yaml file: %w", err)
	}

	fmt.Printf("Rollup config written to: %s\n", filename)
	return nil
}

func getL1RpcUrl(superchainTarget string) (string, error) {
	superChain, ok := superchain.Superchains[superchainTarget]
	if !ok {
		return "", fmt.Errorf("unknown superchain target provided: %s", superchainTarget)
	}

	if superChain.Config.L1.PublicRPC == "" {
		return "", fmt.Errorf("missing L1 public rpc endpoint in superchain config")
	}

	fmt.Printf("Setting L1 public rpc endpoint to %s\n", superChain.Config.L1.PublicRPC)
	return superChain.Config.L1.PublicRPC, nil
}
