package genesis

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
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
	vmd := chain.ValidationMetadata

	if vmd == nil {
		t.Skip("WARNING: cannot yet validate this chain (no validation metadata)")
	}

	monorepoCommit := vmd.GenesisCreationCommit

	// Setup some directory references
	thisDir := getDirOfThisFile()
	monorepoDir := path.Join(thisDir, "../../../optimism-temporary")
	contractsDir := path.Join(monorepoDir, "packages/contracts-bedrock")

	// reset to appropriate commit, this is preferred to git checkout because it will
	// blow away any leftover files from the previous run
	executeCommandInDir(t, monorepoDir, exec.Command("git", "reset", "--hard", monorepoCommit))

	// TODO unskip these, I am skipping to save time in development since we
	// are not validating multiple chains yet
	if false {
		executeCommandInDir(t, monorepoDir, exec.Command("rm", "-rf", "node_modules"))
		executeCommandInDir(t, contractsDir, exec.Command("rm", "-rf", "node_modules"))
	}

	// install dependencies
	// TODO we expect this step to vary as we scan through the monorepo history
	// so we will need some branching logic here
	executeCommandInDir(t, contractsDir, exec.Command("pnpm", "install"))

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

	// regenerate genesis.json  at this monorepo commit.
	executeCommandInDir(t, thisDir, exec.Command("sh", "monorepo-outputs.sh"))

	expectedData, err := os.ReadFile(path.Join(monorepoDir, "expected-genesis.json"))
	require.NoError(t, err)

	g, err := core.LoadOPStackGenesis(chainId)
	require.NoError(t, err)

	gotData, err := g.MarshalJSON()
	require.NoError(t, err)

	os.WriteFile(path.Join(monorepoDir, "got-genesis.json"), gotData, 0777)

	require.Equal(t, string(expectedData), string(gotData))
}

func getDirOfThisFile() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	return filepath.Dir(filename)
}
