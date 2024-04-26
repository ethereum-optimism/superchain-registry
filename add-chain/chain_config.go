package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"gopkg.in/yaml.v3"
)

type RollupConfig struct {
	Name            string      `yaml:"name"`
	L2ChainID       uint64      `json:"l2_chain_id" yaml:"chain_id"`
	PublicRPC       string      `yaml:"public_rpc"`
	SequencerRPC    string      `yaml:"sequencer_rpc"`
	Explorer        string      `yaml:"explorer"`
	SuperchainLevel int         `yaml:"superchain_level"`
	BatchInboxAddr  string      `json:"batch_inbox_address" yaml:"batch_inbox_addr"`
	Genesis         GenesisData `json:"genesis" yaml:"genesis"`
	CanyonTime      *int        `json:"canyon_time,omitempty" yaml:"canyon_time,omitempty"`
	DeltaTime       *int        `json:"delta_time,omitempty" yaml:"delta_time,omitempty"`
	EcotoneTime     *int        `json:"ecotone_time,omitempty" yaml:"ecotone_time,omitempty"`
}

type GenesisData struct {
	L1           GenesisLayer `json:"l1" yaml:"l1"`
	L2           GenesisLayer `json:"l2" yaml:"l2"`
	L2Time       int          `json:"l2_time" yaml:"l2_time"`
	SystemConfig SystemConfig `json:"system_config" yaml:"system_config,omitempty"`
}

type SystemConfig struct {
	BatcherAddr       string `json:"batcherAddr"`
	Overhead          string `json:"overhead"`
	Scalar            string `json:"scalar"`
	GasLimit          uint64 `json:"gasLimit"`
	BaseFeeScalar     uint64 `json:"baseFeeScalar"`
	BlobBaseFeeScalar uint64 `json:"blobBaseFeeScalar"`
}

type GenesisLayer struct {
	Hash   string `json:"hash" yaml:"hash"`
	Number int    `json:"number" yaml:"number"`
}

func constructRollupConfig(filePath, chainName, publicRPC, sequencerRPC, explorer string, superchainLevel int) (RollupConfig, error) {
	fmt.Printf("Attempting to read from %s\n", filePath)
	file, err := os.ReadFile(filePath)
	if err != nil {
		return RollupConfig{}, fmt.Errorf("error reading file: %v", err)
	}
	var config RollupConfig
	if err = json.Unmarshal(file, &config); err != nil {
		return RollupConfig{}, fmt.Errorf("error unmarshaling json: %v", err)
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
	superchainLevel int,
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

	rollupConfig.Genesis.SystemConfig = SystemConfig{} // remove SystemConfig so its omitted from yaml
	yamlData, err := yaml.Marshal(rollupConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}

	// Unmarshal bytes into a yaml.Node for custom manipulation
	var rootNode yaml.Node
	if err = yaml.Unmarshal(yamlData, &rootNode); err != nil {
		return err
	}

	enhanceYAML(&rootNode)

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

func enhanceYAML(node *yaml.Node) {
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0] // Dive into the document node
	}

	var lastKey string
	for i := 0; i < len(node.Content)-1; i += 2 {
		currentNode := node.Content[i]

		// Add blank line AFTER these keys
		if lastKey == "explorer" || lastKey == "superchain_level" || lastKey == "genesis" {
			currentNode.HeadComment = "\n"
		}

		// Add blank line BEFORE these keys
		if currentNode.Value == "genesis" {
			currentNode.HeadComment = "\n"
		}

		lastKey = currentNode.Value
	}
}
