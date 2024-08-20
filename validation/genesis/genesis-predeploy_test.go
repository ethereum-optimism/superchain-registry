package genesis

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/core"
	"github.com/stretchr/testify/require"
)

// TODO deduplicate this
// perChainTestName ensures test can easily be filtered by chain name or chain id using the -run=regex testflag.
func perChainTestName(chain *superchain.ChainConfig) string {
	return chain.Name + fmt.Sprintf(" (%d)", chain.ChainID)
}

func TestGenesisPredeploys(t *testing.T) {
	for _, chain := range OPChains {
		if chain.SuperchainLevel == Standard || chain.StandardChainCandidate {
			t.Run(perChainTestName(chain), func(t *testing.T) {
				// Do not run in parallel
				testGenesisPredeploys(t, chain)
			})
		}
	}
}

// Invoke this with go test -timeout 0 ./validation/genesis -run=TestGenesisPredeploys -v
// REQUIREMENTS:
// pnpm and yarn, so we can prepare https://codeload.github.com/Saw-mon-and-Natalie/clones-with-immutable-args/tar.gz/105efee1b9127ed7f6fedf139e1fc796ce8791f2
func testGenesisPredeploys(t *testing.T, chain *ChainConfig) {
	chainId := chain.ChainID

	vis, ok := ValidationInputs[chainId]

	if !ok {
		t.Skip("WARNING: cannot yet validate this chain (no validation metadata)")

	}

	monorepoCommit := vis.GenesisCreationCommit

	// Setup some directory references
	thisDir := getDirOfThisFile()
	chainIdString := strconv.Itoa(int(chainId))
	validationInputsDir := path.Join(thisDir, "validation-inputs", chainIdString)
	monorepoDir := path.Join(thisDir, "../../../optimism-temporary")
	contractsDir := path.Join(monorepoDir, "packages/contracts-bedrock")

	// reset to appropriate commit, this is preferred to git checkout because it will
	// blow away any leftover files from the previous run
	executeCommandInDir(t, monorepoDir, exec.Command("git", "reset", "--hard", monorepoCommit))

	executeCommandInDir(t, monorepoDir, exec.Command("rm", "-rf", "node_modules"))
	executeCommandInDir(t, contractsDir, exec.Command("rm", "-rf", "node_modules"))


	if monorepoCommit == "d80c145e0acf23a49c6a6588524f57e32e33b91" {
		// apply a patch to get things working
		// then compile the contracts
		// TODO not sure why this is needed, it is likely coupled to the specific commit we are looking at
		executeCommandInDir(t, thisDir, exec.Command("cp", "foundry-config.patch", contractsDir))
		executeCommandInDir(t, contractsDir, exec.Command("git", "apply", "foundry-config.patch"))
		executeCommandInDir(t, contractsDir, exec.Command("forge", "build"))
		// revert patch, makes rerunning script locally easier
		executeCommandInDir(t, contractsDir, exec.Command("git", "apply", "-R", "foundry-config.patch"))
	}

	// copy genesis input files to monorepo
	executeCommandInDir(t, validationInputsDir,
		exec.Command("cp", "deploy-config.json", path.Join(contractsDir, "deploy-config", chainIdString+".json")))
	err := os.MkdirAll(path.Join(contractsDir, "deployments", chainIdString), os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	err = writeDeployments(chainId, path.Join(contractsDir, "deployments", chainIdString))
	if err != nil {
		log.Fatalf("Failed to write deployments: %v", err)
	}
	// writeDeploymentsLegacy(chainId, path.Join(contractsDir, "deployments", chainIdString))

	// regenerate genesis.json at this monorepo commit.
	executeCommandInDir(t, thisDir, exec.Command("cp", "./monorepo-outputs.sh", monorepoDir))
	executeCommandInDir(t, monorepoDir, exec.Command("sh", "./monorepo-outputs.sh", vis.MonorepoBuildCommand, vis.GenesisCreationCommand))

	expectedData, err := os.ReadFile(path.Join(monorepoDir, "expected-genesis.json"))
	require.NoError(t, err)

	gen := core.Genesis{}

	err = json.Unmarshal(expectedData, &gen)
	require.NoError(t, err)

	expectedData, err = json.Marshal(gen.Alloc)
	require.NoError(t, err)

	g, err := core.LoadOPStackGenesis(chainId)
	require.NoError(t, err)

	gotData, err := json.Marshal(g.Alloc)
	require.NoError(t, err)

	os.WriteFile(path.Join(monorepoDir, "want-alloc.json"), expectedData, 0777)
	os.WriteFile(path.Join(monorepoDir, "got-alloc.json"), gotData, 0777)

	require.Equal(t, string(expectedData), string(gotData))
}

func getDirOfThisFile() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	return filepath.Dir(filename)
}

func writeDeployments(chainId uint64, directory string) error {
	as := Addresses[chainId]

	data, err := json.Marshal(as)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(directory, ".deploy"), data, 0777)
	if err != nil {
		return err
	}
	return nil
}

func writeDeploymentsLegacy(chainId uint64, directory string) error {

	// Initialize your struct with some data
	data := Addresses[chainId]

	// Get the reflection value object
	val := reflect.ValueOf(*data)
	typ := reflect.TypeOf(*data)

	// Iterate over the struct fields
	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name      // Get the field name
		fieldValue := val.Field(i).String() // Get the field value (assuming it's a string)

		// Define the JSON object
		jsonData := map[string]string{
			"address": fieldValue,
		}

		// Convert the map to JSON
		fileContent, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			return fmt.Errorf("Failed to marshal JSON for field %s: %v", fieldName, err)
		}

		// Create a file named after the field name
		fileName := fmt.Sprintf("%s.json", fieldName)
		file, err := os.Create(path.Join(directory, fileName))
		if err != nil {
			return fmt.Errorf("Failed to create file for field %s: %v", fieldName, err)
		}
		defer file.Close()

		// Write the JSON content to the file
		_, err = file.Write(fileContent)
		if err != nil {
			return fmt.Errorf("Failed to write JSON to file for field %s: %v", fieldName, err)
		}

		fmt.Printf("Created file: %s\n", fileName)
	}
	return nil
}
