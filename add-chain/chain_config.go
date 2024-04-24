package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
	CanyonTime      *int        `json:"canyon_time,omitempty" yaml:"canyon_time"`
	DeltaTime       *int        `json:"delta_time,omitempty" yaml:"delta_time"`
	EcotoneTime     *int        `json:"ecotone_time,omitempty" yaml:"ecotone_time"`
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

	// create genesis-system-config data
	// (this is deprecated, users should load this from L1, when available via SystemConfig)
	dirPath := filepath.Join(superchainRepoPath, "superchain", "extra", "genesis-system-configs", superchainTarget)

	// Ensure the directory exists
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

	rollupConfig.Genesis.SystemConfig = SystemConfig{} // remove SystemConfig so its omitted from yaml
	yamlData, err := yaml.Marshal(rollupConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}

	filename := filepath.Join(targetDirectory)
	err = os.WriteFile(filename, yamlData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write yaml file: %w", err)
	}
	fmt.Printf("Rollup config written to: %s\n", filename)

	return nil
}
