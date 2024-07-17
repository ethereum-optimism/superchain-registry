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

// WriteChainConfigTPOML accepts a rollupConfig, formats it, and writes a single output toml
// file which includes the following:
//   - general chain info/config
//   - contract and role addresses
//   - genesis system config
//   - optional feature config info, if activated (e.g. plasma)
func WriteChainConfigTOML(rollupConfig superchain.ChainConfig, targetDirectory string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	comments, err := rollupConfig.GenerateTOMLComments(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate toml comments: %w", err)
	}

	// Marshal the struct to TOML
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(rollupConfig); err != nil {
		return fmt.Errorf("failed to marshal toml: %w", err)
	}

	// Create final content with comments
	var finalContent strings.Builder
	lines := strings.Split(buf.String(), "\n")

	for i, line := range lines {
		splits := strings.Split(line, "=")
		lineKey := strings.TrimSpace(splits[0])
		if len(splits) > 1 && strings.TrimSpace(splits[1]) == "\"0x0000000000000000000000000000000000000000\"" {
			// Skip this line to exclude zero addresses from the output file. Makes the config .toml cleaner
			continue
		}
		if comment, exists := comments[lineKey]; exists {
			finalContent.WriteString(line + " " + comment + "\n")
		} else if i != len(lines)-1 || line != "" {
			// Prevent double empty line at the end of the file
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
