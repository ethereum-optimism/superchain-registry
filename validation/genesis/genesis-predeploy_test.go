package genesis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
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

type GenesisAccountLite struct {
	Storage map[string]string  `json:"storage,omitempty"`
	Balance *superchain.HexBig `json:"balance,omitempty"`
	Nonce   uint64             `json:"nonce,omitempty"`
}

type GenesisLite struct {
	// State data
	Alloc map[string]GenesisAccountLite `json:"alloc"`
}

// Invoke this with go test -timeout 0 ./validation -run=TestGenesisPredeploys -v
// REQUIREMENTS:
// yarn, so we can prepare https://codeload.github.com/Saw-mon-and-Natalie/clones-with-immutable-args/tar.gz/105efee1b9127ed7f6fedf139e1fc796ce8791f2
func TestGenesisPredeploys(t *testing.T) {

	artifactNames := map[string]string{
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30013": "L1BlockNumber",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30007": "L2CrossDomainMessenger",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30017": "OptimismMintableERC721Factory",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30011": "SequencerFeeVault",
		"0x4200000000000000000000000000000000000042": "GovernanceToken",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3000f": "GasPriceOracle",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30014": "L2ERC721Bridge",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3001a": "L1FeeVault",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30019": "BaseFeeVault",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30020": "SchemaRegistry",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30021": "EAS",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30012": "OptimismMintableERC20Factory",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30018": "ProxyAdmin",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30002": "DeployerWhitelist", // Deprecated according to specs
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30015": "L1Block",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30016": "L2ToL1MessagePasser",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30010": "L2StandardBridge",
		"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30000": "LegacyMessagePasser", // Deprecated according to specs
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

	// Generate a genesis.json state dump for OP mainnet at this monorepo commit.
	executeCommandInDir(t, monorepoDir, exec.Command(
		"go", "run", "op-node/cmd/main.go", "genesis", "l2",
		"--deploy-config=./packages/contracts-bedrock/deploy-config/mainnet.json",
		"--outfile.l2=expected-genesis.json",
		"--outfile.rollup=rollup.json",
		"--deployment-dir=./packages/contracts-bedrock/deployments/mainnet",
		"--l1-rpc=https://ethereum-rpc.publicnode.com"))

	data, err := os.ReadFile(path.Join(monorepoDir, "expected-genesis.json"))
	require.NoError(t, err)

	expectedGenesis := new(GenesisLite)
	err = json.Unmarshal(data, expectedGenesis)
	require.NoError(t, err)

	// Validate Allocs
	for address, account := range expectedGenesis.Alloc {

		if !strings.HasPrefix(address, "0x") {
			address = "0x" + address
		}

		if address == "0x4200000000000000000000000000000000000042" {
			// GovernanceToken should NOT be set in any standard chain
			// But it is for OP mainnet, hence the need to special case this one
			// TODO instate a proper check here
			// I think sometimes this reverts to an ordinary proxy when governance token is disabled
			// see https://www.notion.so/oplabs/BHIC-L2-Genesis-Predeploy-Verification-Metal-Mode-Zora-f42ecf4dd0164b4d9303b933d9e98bf6#0a94fd37ca37474d998ab1668e655264
			continue
		}

		// Note that although the spec states this is a one-byte (two hexit) namespace, in fact we are using up to three hexits
		// e.g. 4200000000000000000000000000000000000800 is present in the expected genesis
		// Note also we don't want to skip
		if strings.HasPrefix(address, "0x4200000000000000000000000000000000000") && address != "0x4200000000000000000000000000000000000042" {
			// TODO for now we will skip the proxies themselves (but lets not skip unproxied predeploys here)
			continue
		}

		g, err := superchain.LoadGenesis(chainId)
		require.NoError(t, err)

		if _, ok := g.Alloc[superchain.MustHexToAddress(address)]; !ok {
			t.Fatalf("expected an account at %s, but did not find one", address)
		}

		artifactName, ok := artifactNames[address]

		{ // code validation
			if !ok {
				t.Logf("unimplemented artifact path mapping for %s", address)
				continue
			}

			data, err := os.ReadFile(path.Join(contractsDir, "forge-artifacts", artifactName+".sol", artifactName+".json"))
			require.NoError(t, err)

			cd := new(ContractData)
			err = json.Unmarshal(data, cd)
			require.NoError(t, err)
			wantByteCodeHex := cd.DeployedBytecode.Object

			require.NoError(t, err)

			account := g.Alloc[superchain.MustHexToAddress(address)]
			gotByteCode, err := superchain.LoadContractBytecode(account.CodeHash)
			require.NoError(t, err)

			// TODO check if this is already equal, in which case masking is not necessary
			err = maskBytecode(gotByteCode, cd.DeployedBytecode.ImmutableReferences)
			if err != nil {
				t.Errorf("err masking bytecode for %s, %s", address, err)
			}
			gotByteCodeHex := hexutil.Encode(gotByteCode)

			require.Equal(t, wantByteCodeHex, gotByteCodeHex, "address %s failed validation!", address)

			// Just realised that the Semver universal contract used immutables in the past, making immutables far more prolific (due the semver contract
			// being inherited by many other contracts)
			// These would not be security critical immutables, however, since they can't be changed without modifying the rest of the inherit_ing_ contracts bytecode
			// so our "mask and check" validation approach covers us well.
			t.Logf(address+" code ✅ OK! (%s with %d immutable references)", artifactName, countImmutables(cd.DeployedBytecode.ImmutableReferences))
		}

		{ // balance validation

			wantBalance := account.Balance
			gotBalance := g.Alloc[superchain.MustHexToAddress(address)].Balance

			if wantBalance == nil || (*big.Int)(wantBalance).Cmp(big.NewInt(0)) == 0 {
				if gotBalance != nil && (*big.Int)(wantBalance).Cmp(big.NewInt(0)) != 0 {
					t.Errorf("expected nil or zero balance for account %s (%s), but got nonzero", address, artifactName)
				}
				continue
			}

			if gotBalance == nil {
				t.Errorf("expected non nil balance for account %s (%s), but got nil", address, artifactName)
				continue
			}

			require.Equal(t, wantBalance.String(), gotBalance.String())
		}

	}

}

func countImmutables(irs map[string][]ImmutableReference) int {
	count := 0
	for range irs {
		count++
	}
	return count
}

func maskBytecode(b []byte, immutableReferences map[string][]ImmutableReference) error {
	for _, v := range immutableReferences {
		for _, r := range v {
			for i := r.Start; i < r.Start+r.Length; i++ {
				if i >= len(b) {
					return errors.New("immutable references extend beyond bytecode")
				}
				b[i] = 0
			}
		}
	}
	return nil
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
