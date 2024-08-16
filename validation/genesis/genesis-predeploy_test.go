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
	monorepoCommit := chain.ValidationMetadata.GenesisCreationCommit

	if monorepoCommit == nil {
		t.Skip("WARNING: cannot yet validate this chain (no genesis creation commit)")
	}

	// This maps implementation address to contract name
	// which is sufficient to load the relevant compilation artifact
	// from the monorepo(for the contract in question)
	// This has been built up by reading the optimism specs
	artifactNames := map[string]string{
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

	// Setup some directory references
	thisDir := getDirOfThisFile()
	monorepoDir := path.Join(thisDir, "../../../optimism")
	contractsDir := path.Join(monorepoDir, "packages/contracts-bedrock")

	// checkout appropriate commit
	executeCommandInDir(t, monorepoDir, exec.Command("git", "checkout", *monorepoCommit)) // could use reset --hard to make it easier to run again

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
	executeCommandInDir(t, thisDir, exec.Command("cp", "foundry-config.patch", contractsDir))

	// apply a patch to get things working
	// then compile the contracts
	// TODO not sure why this is needed, it is likely coupled to the specific commit we are looking at
	executeCommandInDir(t, contractsDir, exec.Command("git", "apply", "foundry-config.patch"))
	executeCommandInDir(t, contractsDir, exec.Command("forge", "build"))
	// revert patch, makes rerunning script locally easier
	executeCommandInDir(t, contractsDir, exec.Command("git", "apply", "-R", "foundry-config.patch"))

	// Generate a "synthetic" genesis.json state dump for OP mainnet at this monorepo commit.
	executeCommandInDir(t, thisDir, exec.Command("sh", "monorepo-outputs.sh"))

	data, err := os.ReadFile(path.Join(monorepoDir, "expected-genesis.json"))
	require.NoError(t, err)
	syntheticOPMainnetGenesis := new(GenesisLite)
	err = json.Unmarshal(data, syntheticOPMainnetGenesis)
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

		// Load the genesis for the chain we are validating from the superchain module
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

func getDirOfThisFile() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	return filepath.Dir(filename)
}
