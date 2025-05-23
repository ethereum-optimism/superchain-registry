package deployer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/log"
)

// OpDeployer manages the process of building, setting up, and running the op-deployer
// binary to generate genesis states for L2 chains.
type OpDeployer struct {
	lgr                log.Logger
	l1ContractsRelease string
	inputStatePath     string
	rootDir            string
	deployerVersion    string
}

// NewOpDeployer creates a new OpDeployer instance.
func NewOpDeployer(lgr log.Logger, l1ContractsRelease, stateFilepath, rootDir string) (*OpDeployer, error) {
	if l1ContractsRelease == "" {
		return nil, fmt.Errorf("l1ContractsRelease cannot be empty")
	}

	return &OpDeployer{
		lgr:                lgr,
		l1ContractsRelease: l1ContractsRelease,
		inputStatePath:     stateFilepath,
		rootDir:            rootDir,
	}, nil
}

// BuildDeployerBinary builds the appropriate op-deployer binary for the specified
// contracts release and returns the deployer version.
func (d *OpDeployer) BuildBinary() (string, error) {
	// Normalize the contracts string before lookup in versions.json
	// 1. Remove tag:// prefix if present
	// 2. Remove any -rc.X suffix for version matching
	contractsKey := strings.TrimPrefix(d.l1ContractsRelease, "tag://")
	rcSuffixRegex := regexp.MustCompile(`-rc\.[0-9]+$`)
	contractsKey = rcSuffixRegex.ReplaceAllString(contractsKey, "")

	// Read and parse versions.json
	d.lgr.Info("Reading versions.json")
	versionsPath := filepath.Join(d.rootDir, "ops/internal/deployer/versions.json")
	versionsData, err := os.ReadFile(versionsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read versions.json: %w", err)
	}

	var versions map[string]string
	if err := json.Unmarshal(versionsData, &versions); err != nil {
		return "", fmt.Errorf("failed to parse versions.json: %w", err)
	}

	deployerVersion, ok := versions[contractsKey]
	if !ok {
		return "", fmt.Errorf("no deployer version found for contracts: %s", contractsKey)
	}
	d.lgr.Info("Found deployer version", "version", deployerVersion)
	d.deployerVersion = deployerVersion

	// Run the build-deployer script with the appropriate deployer version
	d.lgr.Info("Running build-deployer script")
	scriptPath := filepath.Join(d.rootDir, "scripts", "build-deployer.sh")
	cmd := exec.Command(scriptPath, deployerVersion)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run build-deployer script: %w", err)
	}

	return deployerVersion, nil
}

func (d *OpDeployer) getBinaryPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	deployerPath := filepath.Join(homeDir, ".cache/binaries", d.deployerVersion, "op-deployer")
	if _, err := os.Stat(deployerPath); os.IsNotExist(err) {
		return "", fmt.Errorf("deployer binary not found at path: %s", deployerPath)
	}
	return deployerPath, nil

}

// GetDeployerVersion returns the deployer version.

// SetupStateAndIntent prepares the deployer environment by creating merged state and intent files
// in the specified working directory.
func (d *OpDeployer) SetupStateAndIntent(workdir string) error {
	// Read the state file
	state, err := ReadOpaqueMappingFile(d.inputStatePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	// Determine version based on deployer version and call appropriate merge function
	var mergedIntent, mergedState OpaqueMapping
	// Parse deployer version to determine which merge function to use
	if strings.Contains(d.deployerVersion, ".3.") {
		mergedIntent, mergedState, err = MergeStateV3(state)
	} else {
		mergedIntent, mergedState, err = MergeStateV2(state)
	}
	if err != nil {
		return fmt.Errorf("failed to merge state: %w", err)
	}

	// Write state.json in the temp directory
	stateJSON, err := json.MarshalIndent(mergedState, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state to JSON: %w", err)
	}
	stateTempPath := filepath.Join(workdir, "state.json")
	if err := os.WriteFile(stateTempPath, stateJSON, 0o644); err != nil {
		return fmt.Errorf("failed to write state to temp file: %w", err)
	}
	d.lgr.Info("Wrote state to temporary file", "path", stateTempPath)

	// Write intent.toml in the temp directory
	intentTOML, err := toml.Marshal(mergedIntent)
	if err != nil {
		return fmt.Errorf("failed to marshal intent to TOML: %w", err)
	}
	intentTempPath := filepath.Join(workdir, "intent.toml")
	if err := os.WriteFile(intentTempPath, intentTOML, 0o644); err != nil {
		return fmt.Errorf("failed to write intent to temp file: %w", err)
	}
	d.lgr.Info("Wrote intent to temporary file", "path", intentTempPath)

	return nil
}

// GenerateGenesis runs op-deployer binary to generate a genesis
func (d *OpDeployer) GenerateGenesis(workdir string) (*core.Genesis, error) {
	deployerPath, err := d.getBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get deployer binary path: %w", err)
	}

	// Run `op-deployer apply` to generate the expected genesis
	d.lgr.Info("Running `op-deployer apply`")
	cmd := exec.Command(deployerPath, "apply", "--workdir", workdir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run op-deployer apply: %w", err)
	}

	// Run `op-deployer inspect genesis` to read the expected genesis
	d.lgr.Info("Running `op-deployer inspect genesis`")
	cmd = exec.Command(deployerPath, "inspect", "genesis", "--workdir", workdir)

	output, err := cmd.Output()
	if err != nil {

		return nil, fmt.Errorf("failed to run op-deployer inspect genesis: %w", err)
	}

	var genesis core.Genesis
	if err := json.Unmarshal(output, &genesis); err != nil {
		return nil, fmt.Errorf("failed to parse op-deployer inspect genesis output: %w", err)
	}

	return &genesis, nil
}

func (d *OpDeployer) InspectGenesis(statePath, chainId string) (*core.Genesis, error) {
	// Run `op-deployer inspect genesis` to read the expected genesis
	d.lgr.Info("Running `op-deployer inspect genesis`")

	deployerPath, err := d.getBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get deployer binary path: %w", err)
	}

	state, err := ReadOpaqueMappingFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Write state.json in the temp directory
	stateJSON, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state to JSON: %w", err)
	}

	workdir, err := os.MkdirTemp("", "op-deployer")
	defer os.RemoveAll(workdir)

	stateTempPath := filepath.Join(workdir, "state.json")
	if err := os.WriteFile(stateTempPath, stateJSON, 0o644); err != nil {

		return nil, fmt.Errorf("failed to write state to temp file: %w", err)
	}
	d.lgr.Info("Wrote state to temporary file", "path", stateTempPath)

	cmd := exec.Command(deployerPath, "inspect", "genesis", "--workdir", workdir, chainId)
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		d.lgr.Error(stderr.ReadString(0))
		return nil, fmt.Errorf("failed to run op-deployer inspect genesis: %w", err)
	}

	var genesis core.Genesis
	if err := json.Unmarshal(output, &genesis); err != nil {
		return nil, fmt.Errorf("failed to parse op-deployer inspect genesis output: %w", err)
	}

	return &genesis, nil
}

func (d *OpDeployer) InspectRollup(statePath, chainId string) (*rollup.Config, error) {
	// Run `op-deployer inspect rollup` to read the expected rollup config
	d.lgr.Info("Running `op-deployer inspect rollup`")

	deployerPath, err := d.getBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get deployer binary path: %w", err)
	}

	state, err := ReadOpaqueMappingFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Write state.json in the temp directory
	stateJSON, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state to JSON: %w", err)
	}

	workdir, err := os.MkdirTemp("", "op-deployer")
	defer os.RemoveAll(workdir)

	stateTempPath := filepath.Join(workdir, "state.json")
	if err := os.WriteFile(stateTempPath, stateJSON, 0o644); err != nil {

		return nil, fmt.Errorf("failed to write state to temp file: %w", err)
	}
	d.lgr.Info("Wrote state to temporary file", "path", stateTempPath)

	cmd := exec.Command(deployerPath, "inspect", "rollup", "--workdir", workdir, chainId)
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		d.lgr.Error(stderr.ReadString(0))
		return nil, fmt.Errorf("failed to run op-deployer inspect genesis: %w", err)
	}

	var rollup rollup.Config
	if err := json.Unmarshal(output, &rollup); err != nil {
		return nil, fmt.Errorf("failed to parse op-deployer inspect rollup output: %w", err)
	}

	return &rollup, nil
}
