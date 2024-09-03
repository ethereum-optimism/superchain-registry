package genesis

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	. "github.com/ethereum-optimism/superchain-registry/validation/common"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

var temporaryOptimismDir string

// TestMain is the entry point for testing in this package.
func TestMain(m *testing.M) {
	// Clone optimism into gitignored temporary directory (if that directory does not yet exist)
	// We avoid cloning under the superchain-registry tree, since this causes dependency resolution problems
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	thisDir := filepath.Dir(filename)
	temporaryOptimismDir = path.Join(thisDir, "../../../optimism-temporary")

	// Clone the repo if it the folder doesn't exist
	_, err := os.Stat(temporaryOptimismDir)
	needToClone := os.IsNotExist(err)
	if needToClone {
		mustExecuteCommandInDir(thisDir,
			exec.Command("git", "clone", "--recurse-submodules", "https://github.com/ethereum-optimism/optimism.git", temporaryOptimismDir))
	}

	// Run tests
	exitVal := m.Run()

	// Teardown code:
	// Only if we cloned the directory, now delete it
	// This means during local development, one can clone the
	// repo manually before running the test to speed up runs.
	if needToClone {
		if err := os.RemoveAll(temporaryOptimismDir); err != nil {
			panic("Failed to remove temp directory: " + err.Error())
		}
	}

	// Exit with the result of the tests
	os.Exit(exitVal)
}

func TestGenesisAllocs(t *testing.T) {
	for _, chain := range OPChains {
		if chain.SuperchainLevel == Standard || chain.StandardChainCandidate {
			t.Run(PerChainTestName(chain), func(t *testing.T) {
				// Do not run in parallel, because
				// the sub tests share the temporaryOptimismDir
				// as a resource in a concurrency unsafe way.
				// Parallelism is handled by the CI configuration.
				testGenesisAllocs(t, chain)
			})
		}
	}
}

func testGenesisAllocs(t *testing.T, chain *ChainConfig) {
	chainId := chain.ChainID
	vis, ok := ValidationInputs[chainId]

	if !ok {
		t.Fatalf("Could not validate the genesis of chain %d (no validation metadata)", chainId)
	}

	monorepoCommit := vis.GenesisCreationCommit

	// Setup some directory references
	thisDir := getDirOfThisFile()
	chainIdString := strconv.Itoa(int(chainId))
	validationInputsDir := path.Join(thisDir, "validation-inputs", chainIdString)
	monorepoDir := temporaryOptimismDir
	contractsDir := path.Join(monorepoDir, "packages/contracts-bedrock")

	// This is preferred to git checkout because it will
	// blow away any leftover files from the previous run
	t.Logf("🛠️ Resetting monorepo to %s...", monorepoCommit)
	mustExecuteCommandInDir(monorepoDir, exec.Command("git", "reset", "--hard", monorepoCommit))
	mustExecuteCommandInDir(monorepoDir, exec.Command("git", "submodule", "update"))

	t.Log("🛠️ Deleting node_modules...")
	mustExecuteCommandInDir(monorepoDir, exec.Command("rm", "-rf", "node_modules"))
	mustExecuteCommandInDir(contractsDir, exec.Command("rm", "-rf", "node_modules"))

	t.Log("🛠️ Attempting to apply config.patch...")
	mustExecuteCommandInDir(thisDir, exec.Command("cp", "config.patch", monorepoDir))
	_ = executeCommandInDir(monorepoDir, exec.Command("git", "apply", "config.patch")) // continue on error

	t.Log("🛠️ Copying deploy-config, deployments, and wrapper script to temporary dir...")
	mustExecuteCommandInDir(validationInputsDir,
		exec.Command("cp", "deploy-config.json", path.Join(contractsDir, "deploy-config", chainIdString+".json")))
	err := os.MkdirAll(path.Join(contractsDir, "deployments", chainIdString), os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	if vis.GenesisCreationCommand == "opnode1" {
		err = writeDeploymentsLegacy(chainId, path.Join(contractsDir, "deployments", chainIdString))
	} else {
		err = writeDeployments(chainId, path.Join(contractsDir, "deployments", chainIdString))
	}
	if err != nil {
		log.Fatalf("Failed to write deployments: %v", err)
	}

	var runDir string
	if strings.HasPrefix(vis.GenesisCreationCommand, "forge") {
		runDir = contractsDir
	} else {
		runDir = monorepoDir
	}

	mustExecuteCommandInDir(thisDir, exec.Command("cp", "./monorepo-outputs.sh", runDir))
	buildCommand := BuildCommand[vis.MonorepoBuildCommand]
	if vis.NodeVersion == "" {
		panic("must set node_version in meta.toml")
	}
	creationCommand := GenesisCreationCommand[vis.GenesisCreationCommand](chainId, Superchains[chain.Superchain].Config.L1.PublicRPC)
	cmd := exec.Command("bash", "./monorepo-outputs.sh", vis.NodeVersion, buildCommand, creationCommand)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to get stderr pipe: %v", err)
	}
	// Stream the command's stdout and stderr to the test logger
	go streamOutputToLogger(stdoutPipe, t)
	go streamOutputToLogger(stderrPipe, t)

	t.Log("🛠️ Regenerating genesis...")
	mustExecuteCommandInDir(runDir, cmd)

	t.Log("🛠️ Comparing registry genesis.alloc with regenerated genesis.alloc...")
	var expectedData []byte

	if strings.HasPrefix(vis.GenesisCreationCommand, "forge") {
		expectedData, err = os.ReadFile(path.Join(contractsDir, "statedump.json"))
		require.NoError(t, err)
		allocs := types.GenesisAlloc{}
		err = json.Unmarshal(expectedData, &allocs)
		require.NoError(t, err)
		expectedData, err = json.MarshalIndent(allocs, "", " ")
		require.NoError(t, err)
	} else {
		expectedData, err = os.ReadFile(path.Join(monorepoDir, "expected-genesis.json"))
		require.NoError(t, err)
		gen := core.Genesis{}
		err = json.Unmarshal(expectedData, &gen)
		require.NoError(t, err)
		expectedData, err = json.MarshalIndent(gen.Alloc, "", " ")
		require.NoError(t, err)
	}

	g, err := core.LoadOPStackGenesis(chainId)
	require.NoError(t, err)

	gotData, err := json.MarshalIndent(g.Alloc, "", " ")
	require.NoError(t, err)

	err = os.WriteFile(path.Join(monorepoDir, "want-alloc.json"), expectedData, os.ModePerm) // regenerated
	require.NoError(t, err)
	err = os.WriteFile(path.Join(monorepoDir, "got-alloc.json"), gotData, os.ModePerm) // read from registry
	require.NoError(t, err)

	require.Equal(t, string(expectedData), string(gotData))
}

func getDirOfThisFile() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	return filepath.Dir(filename)
}
