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

	"github.com/ethereum-optimism/superchain-registry/superchain"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/common"
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
	mustExecuteCommandInDir(t, monorepoDir, exec.Command("git", "reset", "--hard", monorepoCommit))
	mustExecuteCommandInDir(t, monorepoDir, exec.Command("rm", "-rf", "node_modules"))
	mustExecuteCommandInDir(t, contractsDir, exec.Command("rm", "-rf", "node_modules"))

	// attempt to apply config.patch
	mustExecuteCommandInDir(t, thisDir, exec.Command("cp", "config.patch", monorepoDir))
	_ = executeCommandInDir(t, monorepoDir, exec.Command("git", "apply", "config.patch")) // continue on error

	// copy genesis input files to monorepo
	mustExecuteCommandInDir(t, validationInputsDir,
		exec.Command("cp", "deploy-config.json", path.Join(contractsDir, "deploy-config", chainIdString+".json")))
	err := os.MkdirAll(path.Join(contractsDir, "deployments", chainIdString), os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	if vis.GenesisCreationCommand == "opnode2" {
		err = writeDeploymentsLegacy(chainId, path.Join(contractsDir, "deployments", chainIdString))
	} else {
		err = writeDeployments(chainId, path.Join(contractsDir, "deployments", chainIdString))
	}
	if err != nil {
		log.Fatalf("Failed to write deployments: %v", err)
	}

	// regenerate genesis.json at this monorepo commit.
	mustExecuteCommandInDir(t, thisDir, exec.Command("cp", "./monorepo-outputs.sh", monorepoDir))
	buildCommand := BuildCommand[vis.MonorepoBuildCommand]
	if vis.NodeVersion == "" {
		panic("must set node_version in meta.toml")
	}
	creationCommand := GenesisCreationCommand[vis.GenesisCreationCommand](chainId, Superchains[chain.Superchain].Config.L1.PublicRPC)
	mustExecuteCommandInDir(t, monorepoDir, exec.Command("sh", "./monorepo-outputs.sh", vis.NodeVersion, buildCommand, creationCommand))

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
	// Prepare a HardHat Deployment type, we need this whole structure to make things
	// work, although it is only the Address field which ends up getting used.
	type StorageLayoutEntry struct {
		AstId    uint   `json:"astId"`
		Contract string `json:"contract"`
		Label    string `json:"label"`
		Offset   uint   `json:"offset"`
		Slot     uint   `json:"slot,string"`
		Type     string `json:"type"`
	}
	type StorageLayoutType struct {
		Encoding      string `json:"encoding"`
		Label         string `json:"label"`
		NumberOfBytes uint   `json:"numberOfBytes,string"`
		Key           string `json:"key,omitempty"`
		Value         string `json:"value,omitempty"`
		Base          string `json:"base,omitempty"`
	}
	type StorageLayout struct {
		Storage []StorageLayoutEntry         `json:"storage"`
		Types   map[string]StorageLayoutType `json:"types"`
	}
	type Deployment struct {
		Name             string
		Abi              []string        `json:"abi"`
		Address          string          `json:"address"`
		Args             []any           `json:"args"`
		Bytecode         string          `json:"bytecode"`
		DeployedBytecode string          `json:"deployedBytecode"`
		Devdoc           json.RawMessage `json:"devdoc"`
		Metadata         string          `json:"metadata"`
		Receipt          json.RawMessage `json:"receipt"`
		SolcInputHash    string          `json:"solcInputHash"`
		StorageLayout    StorageLayout   `json:"storageLayout"`
		TransactionHash  common.Hash     `json:"transactionHash"`
		Userdoc          json.RawMessage `json:"userdoc"`
	}

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
		text, err := v.MarshalText()
		if err != nil || !strings.HasPrefix(string(text), "0x") {
			continue
		}
		// Define the Deployment object, filling in only what we need
		jsonData := Deployment{Address: v.String(), Name: k}

		raw, err := json.MarshalIndent(jsonData, "", " ")
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
