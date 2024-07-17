package validation

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/superchain-registry/superchain"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"

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
	results := make(map[string]map[string]superchain.Hash)

	for _, chain := range OPChains {
		chain := chain
		t.Run(perChainTestName(chain), func(t *testing.T) {
			if chain.StandardChainCandidate == false && chain.SuperchainLevel == Frontier {
				t.Skip()
			}
			if chain.ChainID == 10 {
				t.Skip("we don't have the allocs for op mainnet")
			}

			g, err := LoadGenesis(chain.ChainID)
			require.NoError(t, err)

			stringChainID := strconv.FormatUint(chain.ChainID, 10)
			results[stringChainID] = make(map[string]superchain.Hash)
			for address, account := range g.Alloc {

				if !strings.HasPrefix(address.String(), "0xc0D3C0d3") {
					continue // concentrating on predeploy implementations for now
				}

				results[stringChainID][address.String()] = account.CodeHash
			}
		})
	}

	t.Log(results)
	// Open the file for writing in the current directory
	file, err := os.Create("results.toml")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Encode the struct to TOML and write to the file
	if err := toml.NewEncoder(file).Encode(results); err != nil {
		fmt.Println("Error encoding TOML:", err)
	}
}
