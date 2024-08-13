package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// Define a struct to represent the structure of the JSON data
type DeployedBytecode struct {
	Object string `json:"object"`
}

type ContractData struct {
	DeployedBytecode DeployedBytecode `json:"deployedBytecode"`
}

// REQUIREMENTS:
// yarn, so we can prepare https://codeload.github.com/Saw-mon-and-Natalie/clones-with-immutable-args/tar.gz/105efee1b9127ed7f6fedf139e1fc796ce8791f2
func TestGenesisPredeploys(t *testing.T) {

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	// Get the directory of the current file
	dir := filepath.Dir(filename)

	monorepoDir := path.Join(dir, "../../optimism")
	contractsDir := path.Join(monorepoDir, "packages/contracts-bedrock")

	// chainId := 34443 // Mode mainnet

	monorepoCommit := "d80c145e0acf23a49c6a6588524f57e32e33b91c"

	executeCommandInDir(t, monorepoDir, exec.Command("git", "checkout", monorepoCommit)) // could use reset --hard to make it easier to run again
	executeCommandInDir(t, monorepoDir, exec.Command("git", "fetch", "--recurse-submodules"))

	// TODO unskip these, I am skipping to save time in development
	// executeCommandInDir(t, monorepoDir, exec.Command("rm", "-rf", "node_modules"))
	// executeCommandInDir(t, contractsDir, exec.Command("rm", "-rf", "node_modules"))

	// possible optimization
	// executeCommandInDir(t, monorepoDir, exec.Command("echo", "'recursive-install=true'", ">>", ".npmrc"))
	// executeCommandInDir(t, contractsDir, exec.Command("pnpm", "install"))

	executeCommandInDir(t, contractsDir, exec.Command("pnpm", "install"))
	executeCommandInDir(t, dir, exec.Command("cp", "foundry-config.patch", contractsDir))
	executeCommandInDir(t, contractsDir, exec.Command("git", "apply", "foundry-config.patch"))

	executeCommandInDir(t, contractsDir, exec.Command("forge", "build"))

	data, err := os.ReadFile(path.Join(contractsDir, "forge-artifacts/BaseFeeVault.sol/BaseFeeVault.json"))
	require.NoError(t, err)

	cd := new(ContractData)
	err = json.Unmarshal(data, cd)
	require.NoError(t, err)
	t.Log(cd)
}

func executeCommandInDir(t *testing.T, dir string, cmd *exec.Cmd) {
	t.Logf("executing %s", cmd.String())
	cmd.Dir = dir
	var outErr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &outErr
	err := cmd.Run()
	if err != nil {
		// error case : status code of command is different from 0
		fmt.Println(outErr.String())
		t.Fatal(err)
	}
}
