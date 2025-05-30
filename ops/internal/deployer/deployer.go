package deployer

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/log"
)

//go:embed versions.json
var versionsJSON []byte

// contractVersions maps contract versions to deployer versions
var contractVersions map[string]string

func init() {
	if err := json.Unmarshal(versionsJSON, &contractVersions); err != nil {
		panic(fmt.Sprintf("failed to parse versions.json: %v", err))
	}
}

// OpDeployer manages the process of building a specific binary of op-deployer,
// then shelling out to that binary for various cli commands
type OpDeployer struct {
	DeployerVersion    string // full version including op-deployer/ prefix
	binaryPath         string
	lgr                log.Logger
	l1ContractsRelease string
}

// NewOpDeployer creates a new OpDeployer instance.
// binariesDir is the directory where the op-deployer binaries are stored,
// each in a subdirectory named after the deployer version as in ~/.cache/op-deployer/v0.0.14/op-deployer
func NewOpDeployer(lgr log.Logger, l1ContractsRelease string, binariesDir string) (*OpDeployer, error) {
	if l1ContractsRelease == "" {
		return nil, fmt.Errorf("l1ContractsRelease cannot be empty")
	}
	if binariesDir == "" {
		return nil, fmt.Errorf("binariesDir cannot be empty")
	}

	opd := OpDeployer{
		lgr:                lgr,
		l1ContractsRelease: l1ContractsRelease,
	}

	err := opd.checkBinary(binariesDir)
	if err != nil {
		return nil, fmt.Errorf("failed binary check: %w", err)
	}

	return &opd, nil
}

// checkBinary checks if the required op-deployer binary exists and is executable
func (d *OpDeployer) checkBinary(binariesDir string) error {
	// Normalize the contracts string before lookup in versions map
	// 1. Remove tag:// prefix if present
	// 2. Remove any -rc.X suffix for version matching
	contractsKey := strings.TrimPrefix(d.l1ContractsRelease, "tag://")
	rcSuffixRegex := regexp.MustCompile(`-rc\.[0-9]+$`)
	contractsKey = rcSuffixRegex.ReplaceAllString(contractsKey, "")

	// Look up deployer version in the embedded map
	deployerVersion, ok := contractVersions[contractsKey]
	if !ok {
		return fmt.Errorf("no deployer version found for contracts: %s", contractsKey)
	}
	d.lgr.Info("Found deployer version", "version", deployerVersion)
	d.DeployerVersion = deployerVersion

	// Strip the op-deployer/ prefix for path construction
	strippedVersion := strings.TrimPrefix(deployerVersion, "op-deployer/")
	binaryPath := filepath.Join(binariesDir, strippedVersion, "op-deployer")

	// Check if the binary exists and is executable
	if info, err := os.Stat(binaryPath); err == nil && info.Mode()&0o111 != 0 {
		d.lgr.Info("Found op-deployer binary", "path", binaryPath)
		d.binaryPath = binaryPath
	} else {
		// Binary doesn't exist or isn't executable
		d.lgr.Error("Required op-deployer binary not found", "version", deployerVersion, "expected_path", binaryPath)
		return fmt.Errorf("op-deployer binary not found at %s", binaryPath)
	}

	return nil
}

// setupStateAndIntent prepares the deployer environment by creating merged state and intent files
// in the specified working directory.
func (d *OpDeployer) setupStateAndIntent(inputStatePath, workdir string) error {
	// Read the state file
	state, err := ReadOpaqueStateFile(inputStatePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	// Get state merge function based on deployer version
	mergeFunc, err := getMergeStateFunc(d.DeployerVersion)
	if err != nil {
		return fmt.Errorf("failed to determine merge function: %w", err)
	}
	mergedIntent, mergedState, err := mergeFunc(state)
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
	useInts(mergedIntent)
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

// GenerateStandardGenesis runs op-deployer binary to generate a genesis
// - l1RpcUrl must match the state's L1 and is required by op-deployer, even though we aren't sending any txs
func (d *OpDeployer) GenerateStandardGenesis(statePath, chainId, l1RpcUrl string) (*OpaqueMap, error) {
	workdir, err := d.copyStateFileToTempDir(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to copy state file to temporary directory: %w", err)
	}
	defer os.RemoveAll(workdir)

	if err := d.setupStateAndIntent(statePath, workdir); err != nil {
		return nil, fmt.Errorf("failed to setup state and intent: %w", err)
	}

	// We don't need a funded account here, because there should not be any txs sent.
	// All contracts should have already been deployed, and the 'apply' command should skip
	// those pipeline steps, then only generate the genesis. Therefore use the first account
	// from the test test ... junk mnemonic:
	privateKeyHex := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	// Run `op-deployer apply` to generate the expected genesis
	d.lgr.Info("Running `op-deployer apply`")
	cmd := exec.Command(d.binaryPath, "apply",
		"--workdir", workdir,
		"--deployment-target", "live",
		"--l1-rpc-url", l1RpcUrl,
		"--private-key", privateKeyHex,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run op-deployer apply: %w", err)
	}

	// Run `op-deployer inspect genesis` to read the expected genesis
	d.lgr.Info("Running `op-deployer inspect genesis`")
	cmd = exec.Command(d.binaryPath, "inspect", "genesis", "--workdir", workdir, chainId)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run op-deployer inspect genesis: %w", err)
	}

	var genesis OpaqueMap
	if err := json.Unmarshal(output, &genesis); err != nil {
		return nil, fmt.Errorf("failed to parse op-deployer inspect genesis output: %w", err)
	}

	return &genesis, nil
}

func (d *OpDeployer) copyStateFileToTempDir(statePath string) (string, error) {
	// Create a temporary directory
	workdir, err := os.MkdirTemp("", "op-deployer")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	state, err := ReadOpaqueStateFile(statePath)
	if err != nil {
		return "", fmt.Errorf("failed to read state file: %w", err)
	}

	// Write state.json in the temp directory
	stateJSON, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal state to JSON: %w", err)
	}

	stateTempPath := filepath.Join(workdir, "state.json")
	if err := os.WriteFile(stateTempPath, stateJSON, 0o644); err != nil {
		return "", fmt.Errorf("failed to write state to temp file: %w", err)
	}

	d.lgr.Info("Copied state file to temporary directory", "path", stateTempPath)
	return workdir, nil
}

func (d *OpDeployer) InspectGenesis(statePath, chainId string) (*OpaqueMap, error) {
	// Run `op-deployer inspect genesis` to read the expected genesis
	d.lgr.Info("Running `op-deployer inspect genesis`")

	workdir, err := d.copyStateFileToTempDir(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to copy state file: %w", err)
	}
	defer os.RemoveAll(workdir)

	cmd := exec.Command(d.binaryPath, "inspect", "genesis", "--workdir", workdir, chainId)
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		d.lgr.Error(stderr.ReadString(0))
		return nil, fmt.Errorf("failed to run op-deployer inspect genesis: %w", err)
	}

	var genesis OpaqueMap
	if err := json.Unmarshal(output, &genesis); err != nil {
		return nil, fmt.Errorf("failed to parse op-deployer inspect genesis output: %w", err)
	}

	return &genesis, nil
}

func (d *OpDeployer) InspectRollup(statePath, chainId string) (*rollup.Config, error) {
	// Run `op-deployer inspect rollup` to read the expected rollup config
	d.lgr.Info("Running `op-deployer inspect rollup`")

	workdir, err := d.copyStateFileToTempDir(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to copy state file to temporary directory: %w", err)
	}
	defer os.RemoveAll(workdir)

	cmd := exec.Command(d.binaryPath, "inspect", "rollup", "--workdir", workdir, chainId)
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

func (d *OpDeployer) InspectDeployConfig(statePath, chainId string) (*genesis.DeployConfig, error) {
	// Run `op-deployer inspect deploy-config` to read the expected deploy config
	d.lgr.Info("Running `op-deployer inspect deploy-config`")

	workdir, err := d.copyStateFileToTempDir(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to copy state file to temporary directory: %w", err)
	}
	defer os.RemoveAll(workdir)

	cmd := exec.Command(d.binaryPath, "inspect", "deploy-config", "--workdir", workdir, chainId)
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		d.lgr.Error(stderr.ReadString(0))
		return nil, fmt.Errorf("failed to run op-deployer inspect genesis: %w", err)
	}

	var deployConfig genesis.DeployConfig
	if err := json.Unmarshal(output, &deployConfig); err != nil {
		return nil, fmt.Errorf("failed to parse op-deployer inspect deploy-config output: %w", err)
	}

	return &deployConfig, nil
}
