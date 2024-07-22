package validation

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/naoina/toml"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func testGenesisHash(t *testing.T, chainID uint64) {
	skipIfExcluded(t, chainID)

	chainConfig, ok := OPChains[chainID]
	if !ok {
		t.Fatalf("no chain with ID %d found", chainID)
	}

	declaredGenesisHash := chainConfig.Genesis.L2.Hash

	// In this step, we call a function from op-geth which utilises
	// superchain.OPChains, superchain.LoadGenesis, superchain.LoadContractBytecode
	// to reconstruct a core.Genesis and compute its block hash.
	// We can then compare that against the declared genesis block hash.
	computedGenesis, err := core.LoadOPStackGenesis(chainID)
	require.NoError(t, err)

	computedGenesisHash := computedGenesis.ToBlock().Hash()

	require.Equal(t, common.Hash(declaredGenesisHash), computedGenesisHash, "chain %d: Genesis block hash must match computed value", chainID)
}

func testGenesisHashAgainstRPC(t *testing.T, chain *ChainConfig) {
	skipIfExcluded(t, chain.ChainID)

	declaredGenesisHash := chain.Genesis.L2.Hash
	rpcEndpoint := chain.PublicRPC

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	blockByNumber := func(blockNumber uint64) (*types.Block, error) {
		return client.BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
	}
	genesisBlock, err := Retry(blockByNumber)(chain.Genesis.L2.Number)
	require.NoError(t, err)

	require.Equal(t, genesisBlock.Hash(), common.Hash(declaredGenesisHash), "Genesis Block Hash declared as %s, but RPC returned %s", declaredGenesisHash, genesisBlock.Hash())
}

func testGenesisPredeploys(t *testing.T, chain *ChainConfig) {
	g, err := LoadGenesis(chain.ChainID)
	require.NoError(t, err)

	for address, listOfAcceptableCodeHashes := range standard.Config.Alloc {
		ch := g.Alloc[MustHexToAddress(address)].CodeHash
		require.Contains(t, listOfAcceptableCodeHashes, ch,
			"account %s: codehash %s not present in list of acceptable codehases %s", address, ch, listOfAcceptableCodeHashes)
	}

}

func TestGPs(t *testing.T) {
	for _, chain := range OPChains {
		chain := chain
		t.Run(perChainTestName(chain), func(t *testing.T) {
			t.Parallel()
			if chain.StandardChainCandidate == false && chain.SuperchainLevel == Frontier {
				t.Skip()
			}
			if chain.ChainID == 10 {
				t.Skip("we don't have the allocs for op mainnet")
			}
			testGenesisPredeploys(t, chain)
		})
	}
}

func TestWritePredeployCodeHashSummary(t *testing.T) {
	// This maps implementation address to codehash to chain identifier
	results := make(map[string]map[string][]string)

	prefix := "0xc0d3c0d3c0d3C0d3c0d3C0D3C0D3c0d3C0D3"
	implementationAddresses := []string{ // from the specs
		prefix + "0000",
		prefix + "0002",
		"0xDeadDeAddeAddEAddeadDEaDDEAdDeaDDeAD0000",
		"0x4200000000000000000000000000000000000006",
		prefix + "0007",
		prefix + "0010",
		prefix + "0011",
		prefix + "0012",
		prefix + "0013",
		prefix + "000F",
		prefix + "0042",
		prefix + "0015",
		prefix + "0016",
		prefix + "0014",
		prefix + "0017",
		prefix + "0018",
		prefix + "0019",
		prefix + "001a",
		prefix + "0020",
		prefix + "0021",
		"0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02",
	}
	for _, implementation := range implementationAddresses {
		results[implementation] = make(map[string][]string)
		for _, chain := range OPChains {
			if chain.StandardChainCandidate == false && chain.SuperchainLevel == Frontier {
				continue
			}
			if chain.ChainID == 10 {
				continue
			}

			g, err := LoadGenesis(chain.ChainID)
			require.NoError(t, err)

			ch := g.Alloc[MustHexToAddress(implementation)].CodeHash

			if results[implementation] == nil {
				results[implementation][ch.String()] = []string{chain.Identifier()}
			} else {
				results[implementation][ch.String()] = append(results[implementation][ch.String()], chain.Identifier())
			}
		}
	}

	// Encode the struct to a buffer
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(results); err != nil {
		fmt.Println("Error encoding TOML:", err)
		return
	}

	names := map[string]string{
		prefix + "0000": "LegacyMessagePasser",
		prefix + "0002": "DeployerWhitelist",
		"0xDeadDeAddeAddEAddeadDEaDDEAdDeaDDeAD0000": "LegacyERC20ETH",
		"0x4200000000000000000000000000000000000006": "WETH9",
		prefix + "0007": "L2CrossDomainMessenger",
		prefix + "0010": "L2StandardBridge",
		prefix + "0011": "SequencerFeeVault",
		prefix + "0012": "OptimismMintableERC20Factory",
		prefix + "0013": "L1BlockNumber",
		prefix + "000F": "GasPriceOracle",
		prefix + "0042": "GovernanceToken",
		prefix + "0015": "L1Block",
		prefix + "0016": "L2ToL1MessagePasser",
		prefix + "0014": "L2ERC721Bridge",
		prefix + "0017": "OptimismMintableERC721Factory",
		prefix + "0018": "ProxyAdmin",
		prefix + "0019": "BaseFeeVault",
		prefix + "001a": "L1FeeVault",
		prefix + "0020": "SchemaRegistry",
		prefix + "0021": "EAS",
		"0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02": "BeaconBlockRoot",
	}
	// Function to add comments to specific lines based on a lookup mapping
	addComments := func(input string, lookup map[string]string) string {
		lines := strings.Split(input, "\n")
		for i, line := range lines {
			for key, comment := range lookup {
				if strings.Contains(line, key) {
					lines[i] = fmt.Sprintf("%s # %s", line, comment)
				}
			}
		}
		return strings.Join(lines, "\n")
	}

	// Add comments to the raw string
	modifiedString := addComments(buf.String(), names)

	// Open the file for writing in the current directory
	file, err := os.Create("results.toml")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Write the modified string to the file
	if _, err := file.WriteString(modifiedString); err != nil {
		fmt.Println("Error writing to file:", err)
	}
}
