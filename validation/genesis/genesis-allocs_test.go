package genesis

import (
	"encoding/json"
	"fmt"
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

	"github.com/ethereum/go-ethereum/common"
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

	// Clone the repo if the folder doesn't exist
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
	mustExecuteCommandInDir(monorepoDir, exec.Command("git", "submodule", "update", "--init", "--recursive"))

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
		removeEmptyStorageSlots(allocs, t)

		require.NoError(t, err)
		expectedData, err = json.MarshalIndent(allocs, "", " ")
		require.NoError(t, err)
	} else {
		expectedData, err = os.ReadFile(path.Join(monorepoDir, "expected-genesis.json"))
		require.NoError(t, err)
		gen := core.Genesis{}
		err = json.Unmarshal(expectedData, &gen)
		removeEmptyStorageSlots(gen.Alloc, t)
		require.NoError(t, err)
		expectedData, err = json.MarshalIndent(gen.Alloc, "", " ")
		require.NoError(t, err)
	}

	g, err := core.LoadOPStackGenesis(chainId)
	removeEmptyStorageSlots(g.Alloc, t)
	require.NoError(t, err)

	if chainId == uint64(1301) {
		delete(g.Alloc, common.HexToAddress("0x1f98431c8ad98523631ae4a59f267346ea31f984"))
		delete(g.Alloc, common.HexToAddress("0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f"))
	}

	// Ink Sepolia
	if chainId == uint64(763373) {
		// forge deployer address
		delete(g.Alloc, common.HexToAddress("0xae0bdc4eeac5e950b67c6819b118761caaf61946"))
		// GovernanceToken predeploy storage slot was set to null, which is functionally equivalent to the 0x000...dead address
		governanceTokenAddress := common.HexToAddress("0x4200000000000000000000000000000000000042")
		storageSlot := common.HexToHash("0x000000000000000000000000000000000000000000000000000000000000000a")
		if _, ok := g.Alloc[governanceTokenAddress].
			Storage[storageSlot]; ok {
			require.Fail(t, "expected GovernanceToken contract storage slot to be null")
		}
		g.Alloc[governanceTokenAddress].
			Storage[storageSlot] = common.HexToHash("0x000000000000000000000000deaddeaddeaddeaddeaddeaddeaddeaddeaddead")
	}

	gotData, err := json.MarshalIndent(g.Alloc, "", " ")
	require.NoError(t, err)

	// special handling of weth9 for op-sepolia at the validated commit
	// We've observed that the contract [metadata hash](https://docs.soliditylang.org/en/latest/metadata.html)
	// does not match for the bytecode that is generated and what's in the superchain-registry.
	// The issue is likely a difference in compiler settings when the contract artifacts were generated and
	// stored in the superchain registry. To account for this, we trim the metadata hash portion of the
	// weth9 contract bytecode before writing to file/comparing the outputs
	// For extra safety, we allow this type of check only at the commit hash we know has this issue.
	// In other instances, the metadata may have optionally been excluded from the bytecode in the registry,
	// in which case we ought to check for a complete match.
	if chainId == uint64(11155420) && monorepoCommit == "ba493e94a25df0f646a040c0899bfd0f4d237c06" {
		t.Log("✂️️ Trimming WETH9 bytecode CBOR hash for OP-Sepolia...")
		expectedData, err = trimWeth9BytecodeMetadataHash(expectedData)
		if err != nil {
			err = fmt.Errorf("Regenerated alloc: %w", err)
		}
		require.NoError(t, err)
		gotData, err = trimWeth9BytecodeMetadataHash(gotData)
		if err != nil {
			err = fmt.Errorf("Registry alloc: %w", err)
		}
		require.NoError(t, err)
	}

	err = os.WriteFile(path.Join(monorepoDir, "regenerated-alloc.json"), expectedData, os.ModePerm) // regenerated
	require.NoError(t, err)
	err = os.WriteFile(path.Join(monorepoDir, "registry-alloc.json"), gotData, os.ModePerm) // read from registry
	require.NoError(t, err)

	require.Equal(t, string(expectedData), string(gotData))
}

// This function removes empty storage slots as we know declaring empty slots is functionally equivalent to not declaring them.
func removeEmptyStorageSlots(allocs types.GenesisAlloc, t *testing.T) {
	for _, account := range allocs {
		for slot, value := range account.Storage {
			if value == (common.Hash{}) {
				delete(account.Storage, slot)
				t.Log("Removed empty storage slot: ", slot.Hex())
			}
		}
	}
}

// trim the CBOR octets from the bytecode of a weth9 contract
func trimWeth9BytecodeMetadataHash(data []byte) ([]byte, error) {
	var weth9Data map[string]interface{}

	err := json.Unmarshal(data, &weth9Data)
	if err != nil {
		return nil, fmt.Errorf("error parsing alloc: %w", err)
	}

	// https://specs.optimism.io/protocol/predeploys.html#weth9
	key := "0x4200000000000000000000000000000000000006"
	if entry, ok := weth9Data[key].(map[string]interface{}); ok {
		if code, ok := entry["code"].(string); ok {
			// Trim the last 100 characters (50 bytes) which is the CBOR length in octets
			if len(code) > 100 {
				entry["code"] = code[:len(code)-100]
			} else {
				return nil, fmt.Errorf("The length is less than 100 characters; no trimming performed.")
			}
		} else {
			return nil, fmt.Errorf("Field 'code' not found or is not a string: %s", key)
		}
	} else {
		return nil, fmt.Errorf("Key %s not found", key)
	}

	data, err = json.Marshal(weth9Data)
	if err != nil {
		return nil, fmt.Errorf("error marshalling: %w", err)
	}

	return data, nil
}

func getDirOfThisFile() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	return filepath.Dir(filename)
}
