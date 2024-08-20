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

	// copy genesis input files to monorepo
	executeCommandInDir(t, validationInputsDir,
		exec.Command("cp", "deploy-config.json", path.Join(contractsDir, "deploy-config", chainIdString+".json")))
	err := os.MkdirAll(path.Join(contractsDir, "deployments", chainIdString), os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	if vis.UseLegacyDeploymentsFormat {
		err = writeDeploymentsLegacy(chainId, path.Join(contractsDir, "deployments", chainIdString))
	} else {
		err = writeDeployments(chainId, path.Join(contractsDir, "deployments", chainIdString))
	}
	if err != nil {
		log.Fatalf("Failed to write deployments: %v", err)
	}

	// regenerate genesis.json at this monorepo commit.
	executeCommandInDir(t, thisDir, exec.Command("cp", "./monorepo-outputs.sh", monorepoDir))
	executeCommandInDir(t, monorepoDir, exec.Command("sh", "./monorepo-outputs.sh", vis.NodeVersion, vis.MonorepoBuildCommand, vis.GenesisCreationCommand))

	expectedData, err := os.ReadFile(path.Join(monorepoDir, "expected-genesis.json"))
	require.NoError(t, err)

	gen := core.Genesis{}

	err = json.Unmarshal(expectedData, &gen)
	require.NoError(t, err)

	expectedData, err = json.MarshalIndent(gen.Alloc, "", " ")
	require.NoError(t, err)

	g, err := core.LoadOPStackGenesis(chainId)
	require.NoError(t, err)

	gotData, err := json.MarshalIndent(g.Alloc, "", " ")
	require.NoError(t, err)

	err = os.WriteFile(path.Join(monorepoDir, "want-alloc.json"), expectedData, 0o777)
	require.NoError(t, err)
	err = os.WriteFile(path.Join(monorepoDir, "got-alloc.json"), gotData, 0o777)
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

func writeDeployments(chainId uint64, directory string) error {
	as := Addresses[chainId]

	data, err := json.Marshal(as)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(directory, ".deploy"), data, 0o777)
	if err != nil {
		return err
	}
	return nil
}

func writeDeploymentsLegacy(chainId uint64, directory string) error {
	// Initialize your struct with some data
	data := Addresses[chainId]

	type AddressList2 AddressList // use another type to prevent infinite recursion later on
	b := AddressList2(*data)

	o, err := json.Marshal(b)
	if err != nil {
		return err
	}

	out := make(map[string]Address)
	err = json.Unmarshal(o, &out)
	if err != nil {
		return err
	}

	for k, v := range out {
		// Define the JSON object
		jsonData := map[string]string{
			"address": v.String(),
		}

		raw, err := json.Marshal(jsonData)
		if err != nil {
			return err
		}

		fileName := fmt.Sprintf("%s.json", k)
		file, err := os.Create(path.Join(directory, fileName))
		if err != nil {
			return fmt.Errorf("Failed to create file for field %s: %w", k, err)
		}
		defer file.Close()

		// Write the JSON content to the file
		_, err = file.Write(raw)
		if err != nil {
			return fmt.Errorf("Failed to write JSON to file for field %s: %w", k, err)
		}

		fmt.Printf("Created file: %s\n", fileName)
	}
	return nil
}
