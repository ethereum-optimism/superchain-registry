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

// constructRollupConfig creates and populates a ChainConfig struct by reading from an input file and
// explicitly setting some additional fields to input argument values
func constructRollupConfig(inputFilePath, chainName, publicRPC, sequencerRPC, explorer string, superchainLevel superchain.SuperchainLevel) (superchain.ChainConfig, error) {
	fmt.Printf("Attempting to read from %s\n", inputFilePath)
	file, err := os.ReadFile(inputFilePath)
	if err != nil {
		return superchain.ChainConfig{}, fmt.Errorf("error reading file: %v", err)
	}
	var config superchain.ChainConfig
	if err = json.Unmarshal(file, &config); err != nil {
		return superchain.ChainConfig{}, fmt.Errorf("error unmarshaling json: %v", err)
	}

	config.Name = chainName
	config.PublicRPC = publicRPC
	config.SequencerRPC = sequencerRPC
	config.SuperchainLevel = superchainLevel
	config.Explorer = explorer

	fmt.Printf("Rollup config successfully constructed\n")
	return config, nil
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

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	systemConfigJSON, err := json.MarshalIndent(rollupConfig.Genesis.SystemConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal genesis system config json: %w", err)
	}

	// Write the genesis system config JSON to a new file
	filePath := filepath.Join(dirPath, rollupConfig.Name+".json")
	if err := os.WriteFile(filePath, systemConfigJSON, 0644); err != nil {
		return fmt.Errorf("failed to write genesis system config json: %w", err)
	}
	fmt.Printf("Genesis system config written to: %s\n", filePath)

	rollupConfig.Genesis.SystemConfig = superchain.SystemConfig{} // remove SystemConfig so its omitted from yaml

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
