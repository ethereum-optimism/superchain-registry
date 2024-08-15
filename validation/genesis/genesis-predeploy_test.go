package genesis

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

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
)

// Define a struct to represent the structure of the JSON data
type DeployedBytecode struct {
	Object              string                          `json:"object"`
	ImmutableReferences map[string][]ImmutableReference `json:"immutableReferences"`
}

type ImmutableReference struct {
	Start  int `json:"start"`
	Length int `json:"length"`
}

type ContractData struct {
	DeployedBytecode DeployedBytecode `json:"deployedBytecode"`
}

// Invoke this with go test -timeout 0 ./validation -run=TestGenesisPredeploys -v
// REQUIREMENTS:
// yarn, so we can prepare https://codeload.github.com/Saw-mon-and-Natalie/clones-with-immutable-args/tar.gz/105efee1b9127ed7f6fedf139e1fc796ce8791f2
func TestGenesisPredeploys(t *testing.T) {

	addressesToCheck := map[string]string{
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3001a": "forge-artifacts/L1FeeVault.sol/L1FeeVault.json",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30019": "forge-artifacts/BaseFeeVault.sol/BaseFeeVault.json",
		// "0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30020": "forge-artifacts/SchemaRegistry.sol/SchemaRegistry.json", This is missing for mode
		// "0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30021": "forge-artifacts/EAS.sol/EAS.json"
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	// Get the directory of the current file
	dir := filepath.Dir(filename)

	monorepoDir := path.Join(dir, "../../../optimism")
	contractsDir := path.Join(monorepoDir, "packages/contracts-bedrock")

	chainId := uint64(34443) // Mode mainnet

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
	executeCommandInDir(t, contractsDir, exec.Command("git", "apply", "-R", "foundry-config.patch")) // revert patch, makes rerunning script locally easier

	for address, artifactPath := range addressesToCheck {
		data, err := os.ReadFile(path.Join(contractsDir, artifactPath))
		require.NoError(t, err)

		cd := new(ContractData)
		err = json.Unmarshal(data, cd)
		require.NoError(t, err)
		wantByteCodeHex := cd.DeployedBytecode.Object

		require.NoError(t, err)

		g, err := superchain.LoadGenesis(chainId)
		require.NoError(t, err)

		account := g.Alloc[superchain.MustHexToAddress(address)]
		gotByteCode, err := superchain.LoadContractBytecode(account.CodeHash)
		require.NoError(t, err)

		maskBytecode(gotByteCode, cd.DeployedBytecode.ImmutableReferences)
		gotByteCodeHex := hexutil.Encode(gotByteCode)

		require.Equal(t, wantByteCodeHex, gotByteCodeHex, "address %s failed validation!", address)
		t.Log(address + " OK!\n")
	}

}

func maskBytecode(b []byte, immutableReferences map[string][]ImmutableReference) {
	for _, v := range immutableReferences {
		for _, r := range v {
			for i := r.Start; i < r.Start+r.Length; i++ {
				b[i] = 0
			}
		}
	}
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
