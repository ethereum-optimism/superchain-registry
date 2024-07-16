package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/superchain-registry/superchain"
	"gopkg.in/yaml.v3"
)

type JSONChainConfig struct {
	ChainID                          uint64                   `json:"l2_chain_id"`
	BatchInboxAddr                   superchain.Address       `json:"batch_inbox_address"`
	Genesis                          superchain.ChainGenesis  `json:"genesis"`
	BlockTime                        uint64                   `json:"block_time"`
	SequencerWindowSize              uint64                   `json:"seq_window_size"`
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

// ConstructChainConfig creates and populates a ChainConfig struct by reading from an input file and
// explicitly setting some additional fields to input argument values
func ConstructChainConfig(
	inputFilePath,
	chainName,
	publicRPC,
	sequencerRPC,
	explorer string,
	superchainLevel superchain.SuperchainLevel,
	standardChainCandidate bool,
	isFaultProofs bool,
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
		Name:                   chainName,
		ChainID:                jsonConfig.ChainID,
		PublicRPC:              publicRPC,
		SequencerRPC:           sequencerRPC,
		Explorer:               explorer,
		BatchInboxAddr:         jsonConfig.BatchInboxAddr,
		Genesis:                jsonConfig.Genesis,
		SuperchainLevel:        superchainLevel,
		StandardChainCandidate: standardChainCandidate,
		SuperchainTime:         nil,
		Plasma:                 jsonConfig.PlasmaConfig,
		HardForkConfiguration: superchain.HardForkConfiguration{
			CanyonTime:  jsonConfig.CanyonTime,
			DeltaTime:   jsonConfig.DeltaTime,
			EcotoneTime: jsonConfig.EcotoneTime,
			FjordTime:   jsonConfig.FjordTime,
		},
		BlockTime:           jsonConfig.BlockTime,
		SequencerWindowSize: jsonConfig.SequencerWindowSize,
	}

	fmt.Printf("Rollup config successfully constructed\n")
	return chainConfig, nil
}

// WriteChainConfig accepts a rollupConfig, formats it, and writes some output files based on the given
// target directories
func WriteChainConfig(
	rollupConfig superchain.ChainConfig,
	targetDirectory string,
) error {
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

func WriteChainConfigTOML(rollupConfig superchain.ChainConfig, targetDirectory string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	comments, err := rollupConfig.GenerateTOMLComments(ctx)
	if err != nil {
		return fmt.Errorf("failed to enhance toml: %w", err)
	}

	// Marshal the struct to TOML
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(rollupConfig); err != nil {
		return fmt.Errorf("failed to marshal toml: %w", err)
	}

	// Create final content with comments
	var finalContent strings.Builder
	for _, line := range strings.Split(buf.String(), "\n") {
		lineKey := strings.Split(line, "=")[0]
		lineKey = strings.TrimSpace(lineKey)
		if comment, exists := comments[lineKey]; exists {
			finalContent.WriteString(line + " " + comment + "\n")
		} else {
			finalContent.WriteString(line + "\n")
		}
	}

	// Write the enhanced TOML data to a file
	filename := filepath.Join(targetDirectory)
	if err := os.WriteFile(filename, []byte(finalContent.String()), 0o644); err != nil {
		return fmt.Errorf("failed to write toml file: %w", err)
	}

	fmt.Printf("Rollup config written to: %s\n", filename)
	return nil
}

func GetL1RpcUrl(superchainTarget string) (string, error) {
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
