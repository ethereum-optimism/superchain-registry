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

type BinaryPicker interface {
	Path() string
	Merger() StateMerger
}

type FixedBinaryPicker struct {
	binaryPath string
	merger     StateMerger
}

func (f *FixedBinaryPicker) Path() string {
	return f.binaryPath
}

func (f *FixedBinaryPicker) Merger() StateMerger {
	return f.merger
}

func WithFixedBinary(binaryPath string, merger StateMerger) *FixedBinaryPicker {
	return &FixedBinaryPicker{
		binaryPath: binaryPath,
		merger:     merger,
	}
}

func WithReleaseBinary(binDir string, l1ContractsRelease string) (*FixedBinaryPicker, error) {
	// Normalize the contracts string before lookup in versions map
	// 1. Remove tag:// prefix if present
	// 2. Remove any -rc.X suffix for version matching
	contractsKey := strings.TrimPrefix(l1ContractsRelease, "tag://")
	rcSuffixRegex := regexp.MustCompile(`-rc\.[0-9]+$`)
	contractsKey = rcSuffixRegex.ReplaceAllString(contractsKey, "")

	// Look up deployer version in the embedded map
	deployerVersion, ok := contractVersions[contractsKey]
	if !ok {
		return nil, fmt.Errorf("no deployer version found for contracts: %s", contractsKey)
	}

	binaryPath := VersionedBinaryPath(binDir, deployerVersion)

	merger, err := GetStateMerger(deployerVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get state merger: %w", err)
	}

	return WithFixedBinary(binaryPath, merger), nil
}

func VersionedBinaryPath(binDir string, deployerVersion string) string {
	return filepath.Join(binDir, fmt.Sprintf("op-deployer_%s", strings.TrimPrefix(deployerVersion, "op-deployer/")))
}

func init() {
	if err := json.Unmarshal(versionsJSON, &contractVersions); err != nil {
		panic(fmt.Sprintf("failed to parse versions.json: %v", err))
	}
}

// OpDeployer manages the process of building a specific binary of op-deployer,
// then shelling out to that binary for various cli commands
type OpDeployer struct {
	binaryPath string
	merger     StateMerger
	lgr        log.Logger
}

// NewOpDeployer creates a new OpDeployer instance.
func NewOpDeployer(lgr log.Logger, binaryPicker BinaryPicker) (*OpDeployer, error) {
	binaryPath := binaryPicker.Path()
	if info, err := os.Stat(binaryPath); err != nil || info.Mode()&0o111 == 0 {
		return nil, fmt.Errorf("op-deployer binary not found or not executable at %s", binaryPath)
	}

	lgr.Info("Using op-deployer binary", "path", binaryPath)

	return &OpDeployer{
		binaryPath: binaryPath,
		merger:     binaryPicker.Merger(),
		lgr:        lgr,
	}, nil
}

// setupStateAndIntent prepares the deployer environment by creating merged state and intent files
// in the specified working directory.
func (d *OpDeployer) setupStateAndIntent(inputStatePath, workdir string) error {
	// Read the state file
	state, err := ReadOpaqueStateFile(inputStatePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	mergedIntent, mergedState, err := d.merger(state)
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

// runCommand executes a command and returns its output, handling stderr capture
func (d *OpDeployer) runCommand(args ...string) ([]byte, error) {
	cmd := exec.Command(d.binaryPath, args...)

	d.lgr.Info("Running command", "command", args[0])
	stderr := new(bytes.Buffer)
	cmd.Stderr = stderr
	output, err := cmd.Output()
	if err != nil {
		d.lgr.Error(stderr.String())
		return nil, fmt.Errorf("failed to run %s: %w", strings.Join(args, " "), err)
	}
	d.lgr.Info("Command success", "command", args[0])
	return output, nil
}

// inspectCommand runs an inspect command and unmarshals the output
func (d *OpDeployer) inspectCommand(workdir, chainId, subcommand string, result interface{}) error {
	output, err := d.runCommand("inspect", subcommand, "--workdir", workdir, chainId)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(output, result); err != nil {
		return fmt.Errorf("failed to parse op-deployer inspect %s output: %w", subcommand, err)
	}
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
	if _, err := d.runCommand("apply",
		"--workdir", workdir,
		"--deployment-target", "live",
		"--l1-rpc-url", l1RpcUrl,
		"--private-key", privateKeyHex); err != nil {
		return nil, err
	}

	// Run `op-deployer inspect genesis` to read the expected genesis
	var genesis OpaqueMap
	if err := d.inspectCommand(workdir, chainId, "genesis", &genesis); err != nil {
		return nil, err
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
	workdir, err := d.copyStateFileToTempDir(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to copy state file: %w", err)
	}
	defer os.RemoveAll(workdir)

	var genesis OpaqueMap
	if err := d.inspectCommand(workdir, chainId, "genesis", &genesis); err != nil {
		return nil, err
	}

	return &genesis, nil
}

func (d *OpDeployer) InspectRollup(statePath, chainId string) (*rollup.Config, error) {
	workdir, err := d.copyStateFileToTempDir(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to copy state file: %w", err)
	}
	defer os.RemoveAll(workdir)

	var rollupConfig rollup.Config
	if err := d.inspectCommand(workdir, chainId, "rollup", &rollupConfig); err != nil {
		return nil, err
	}

	return &rollupConfig, nil
}

func (d *OpDeployer) InspectDeployConfig(statePath, chainId string) (*genesis.DeployConfig, error) {
	workdir, err := d.copyStateFileToTempDir(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to copy state file: %w", err)
	}
	defer os.RemoveAll(workdir)

	var deployConfig genesis.DeployConfig
	if err := d.inspectCommand(workdir, chainId, "deploy-config", &deployConfig); err != nil {
		return nil, err
	}

	return &deployConfig, nil
}
