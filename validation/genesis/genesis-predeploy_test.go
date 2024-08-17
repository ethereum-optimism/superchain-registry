package genesis

import (
	"encoding/json"
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
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

var thisDir, monorepoDir, contractsDir string

// This maps implementation address to contract name
// which is sufficient to load the relevant compilation artifact
// from the monorepo(for the contract in question)
// This has been built up by reading the optimism specs
var predeployArtifactNames = map[string]string{
	"0x4200000000000000000000000000000000000042": "GovernanceToken",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30000": "LegacyMessagePasser", // Deprecated according to specs
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30002": "DeployerWhitelist",   // Deprecated according to specs
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30007": "L2CrossDomainMessenger",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3000f": "GasPriceOracle",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30010": "L2StandardBridge",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30011": "SequencerFeeVault",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30012": "OptimismMintableERC20Factory",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30013": "L1BlockNumber",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30014": "L2ERC721Bridge",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30015": "L1Block",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30016": "L2ToL1MessagePasser",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30017": "OptimismMintableERC721Factory",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30018": "ProxyAdmin",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30019": "BaseFeeVault",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3001a": "L1FeeVault",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30020": "SchemaRegistry",
	"0xc0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d3c0d30021": "EAS",
}

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
	thisDir = getDirOfThisFile()
	monorepoDir = path.Join(thisDir, "../../../optimism-temporary")
	contractsDir = path.Join(monorepoDir, "packages/contracts-bedrock")

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

	// Generate a "synthetic" genesis.json state dump for OP mainnet at this monorepo commit.
	executeCommandInDir(t, thisDir, exec.Command("sh", "monorepo-outputs.sh"))

	data, err := os.ReadFile(path.Join(monorepoDir, "expected-genesis.json"))
	require.NoError(t, err)
	syntheticOPMainnetGenesis := new(GenesisLite)
	err = json.Unmarshal(data, syntheticOPMainnetGenesis)
	require.NoError(t, err)

	// Load the genesis for the chain we are validating from the superchain module
	g, err := LoadGenesis(chainId)
	require.NoError(t, err)

	// From this point on, we are validating the chain's genesis against the synthetic OP Mainnet genesis
	// We used compilation artifacts to help us to this. They help us figure out where the immutable references are
	// So we can mask out the places where we expect the synthetic OP Mainnet genesis to differ from genesis of the chain we
	// are validating.

	// Validate Allocs
	for address, account := range syntheticOPMainnetGenesis.Alloc {

		// I believe this is a legacy format used by geth
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

		if _, ok := g.Alloc[MustHexToAddress(address)]; !ok {
			t.Fatalf("expected an account at %s, but did not find one", address)
		}

		validateCode(t, address, g, syntheticOPMainnetGenesis)
		validateBalance(t, address, account, g)
	}
}

func getDirOfThisFile() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	return filepath.Dir(filename)
}

func validateCode(t *testing.T, address string, g *Genesis, syntheticOPMainnetGenesis *GenesisLite) {
	account := g.Alloc[MustHexToAddress(address)]
	if account.CodeHash == *new(Hash) {
		// no validation needed if codeHash is set to zero value
		return
	}
	gotByteCode, err := LoadContractBytecode(account.CodeHash)
	require.NoError(t, err)

	var wantByteCodeHex string
	artifactName, isPredeploy := predeployArtifactNames[address]
	cd := new(ContractData)
	if isPredeploy {
		// it is a predeploy, we need to perform masking in order to validate it
		filePath := path.Join(contractsDir, "forge-artifacts", artifactName+".sol", artifactName+".json")
		data, err := os.ReadFile(filePath)
		// Check if the error is due to the file not existing
		if err != nil {
			if os.IsNotExist(err) {
				t.Logf("File not found: %s", filePath)
				return
			} else {
				require.NoError(t, err)
			}
		}

		err = json.Unmarshal(data, cd)
		require.NoError(t, err)

		wantByteCodeHex = cd.DeployedBytecode.Object
		err = maskBytecode(gotByteCode, cd.DeployedBytecode.ImmutableReferences)
		require.NoError(t, err, "err masking bytecode for %s, %s", address, err)
	} else {
		// otherwise grab code from synthetic genesis
		wantByteCodeHex = syntheticOPMainnetGenesis.Alloc[address].Code
	}

	wantByteCode, err := hexutil.Decode(wantByteCodeHex)
	require.NoError(t, err)

	if len(gotByteCode) != len(wantByteCode) {
		t.Errorf("expected bytecode at %s to have length %d, but got bytecode with length %d", address, len(wantByteCode), len(gotByteCode))
		return
	}

	// Suppressing this because the output is very verbose. Sometimes useful to pipe into https://difff.jp/en/ though
	if false {
		gotByteCodeHex := hexutil.Encode(gotByteCode)
		require.Equal(t, wantByteCodeHex, gotByteCodeHex, "address %s failed bytecode validation!", address)
	}
	require.Equal(t, crypto.Keccak256Hash(wantByteCode), crypto.Keccak256Hash(gotByteCode), "address %s failed bytecodehash validation!", address)

	if isPredeploy {
		// Just realised that the Semver universal contract used immutables in the past, making immutables far more prolific (due the semver contract
		// being inherited by many other contracts)
		// These would not be security critical immutables, however, since they can't be changed without modifying the rest of the inherit_ing_ contracts bytecode
		// so our "mask and check" validation approach covers us well.
		t.Logf(address+" code ✅ OK! (%s with %d immutable references)", artifactName, countImmutables(cd.DeployedBytecode.ImmutableReferences))
	} else {
		t.Logf(address + " code ✅ OK!")
	}
}

func validateBalance(t *testing.T, address string, account GenesisAccountLite, g *Genesis) {
	wantBalance := account.Balance
	gotBalance := g.Alloc[MustHexToAddress(address)].Balance

	if wantBalance == nil || (*big.Int)(wantBalance).Cmp(big.NewInt(0)) == 0 {
		if gotBalance != nil && (*big.Int)(wantBalance).Cmp(big.NewInt(0)) != 0 {
			t.Errorf("expected nil or zero balance for account %s, but got nonzero", address)
		}
		return
	}

	if gotBalance == nil {
		t.Errorf("expected non nil balance for account %s, but got nil", address)
		return
	}

	require.Equal(t, wantBalance.String(), gotBalance.String())
}
