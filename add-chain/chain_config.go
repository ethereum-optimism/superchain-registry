package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"gopkg.in/yaml.v3"
)

func constructRollupConfig(filePath, chainName, publicRPC, sequencerRPC, explorer string, superchainLevel superchain.SuperchainLevel) (superchain.ChainConfig, error) {
	fmt.Printf("Attempting to read from %s\n", filePath)
	file, err := os.ReadFile(filePath)
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

func writeChainConfig(
	inputFilepath string,
	targetDirectory string,
	chainName string,
	publicRPC string,
	sequencerRPC string,
	explorer string,
	superchainLevel superchain.SuperchainLevel,
	superchainRepoPath string,
	superchainTarget string,
) error {
	rollupConfig, err := constructRollupConfig(inputFilepath, chainName, publicRPC, sequencerRPC, explorer, superchainLevel)
	if err != nil {
		return fmt.Errorf("failed to construct rollup config: %w", err)
	}

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
	filePath := filepath.Join(dirPath, chainName+".json")
	if err := os.WriteFile(filePath, systemConfigJSON, 0644); err != nil {
		return fmt.Errorf("failed to write genesis system config json: %w", err)
	}
	fmt.Printf("Genesis system config written to: %s\n", filePath)

	// Write the rollup config to a yaml file
	filename := filepath.Join(targetDirectory)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

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
	if err = enhanceYAML(ctx, &rootNode); err != nil {
		return err
	}

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

func enhanceYAML(ctx context.Context, node *yaml.Node) error {
	// Check if context is done before processing
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0] // Dive into the document node
	}

	var lastKey string
	for i := 0; i < len(node.Content)-1; i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]

		// Add blank line AFTER these keys
		if lastKey == "explorer" || lastKey == "superchain_level" || lastKey == "genesis" {
			keyNode.HeadComment = "\n"
		}

		// Add blank line BEFORE these keys
		if keyNode.Value == "genesis" {
			keyNode.HeadComment = "\n"
		}

		// Recursive call to check nested fields for "_time" suffix
		if valNode.Kind == yaml.MappingNode {
			if err := enhanceYAML(ctx, valNode); err != nil {
				return err
			}
		}

		// Add human readable timestamp in comment
		if strings.HasSuffix(keyNode.Value, "_time") {
			t, err := strconv.ParseInt(valNode.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to convert yaml string timestamp to int: %w", err)
			}
			timestamp := time.Unix(t, 0)
			keyNode.LineComment = timestamp.Format("Mon 2 Jan 2006 15:04:05 UTC")
		}

		lastKey = keyNode.Value
	}
	return nil
}
