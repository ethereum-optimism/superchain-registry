package validation

import (
	"fmt"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testProxyAdminIsAdminOfPredeploysForChain(t *testing.T, chain ChainConfig) {
	// Create an ethclient connection to the specified RPC URL
	client, err := ethclient.Dial(chain.PublicRPC)
	require.NoError(t, err, "Failed to connect to the Ethereum client at RPC url %s", chain.PublicRPC)
	defer client.Close()
	contractCallResolutions := standard.Config[OPChains[chain.ChainID].Superchain].L2.Universal

	for contract, methodToOutput := range contractCallResolutions {
		for method, output := range methodToOutput {
			t.Run(fmt.Sprintf("%s.%s", contract, method), func(t *testing.T) {
				method := method
				output := output
				t.Parallel()
				want := Address(common.HexToAddress(output))

				got, err := getAddress(method, MustHexToAddress(contract), client)
				require.NoErrorf(t, err, "problem calling %s.%s %s", contract, contract, method)

				assert.Equal(t, want, got, "%s.%s = %s, expected %s", contract, method, got, want)
			})
		}
	}

}

func TestProxyAdminIsAdminOfPredeploys(t *testing.T) {
	isExcluded := map[uint64]bool{
		11155421: true, // sepolia-dev-0/oplabs-devnet-0   No Public RPC declared
		11763072: true, // sepolia-dev-0/base-devnet-0     No Public RPC declared
	}

	for chainID, chain := range OPChains {
		t.Run(perChainTestName(chain), func(t *testing.T) {
			if isExcluded[chainID] {
				t.Skip("chain excluded from check")
			}
			SkipCheckIfFrontierChain(t, *chain)
			testProxyAdminIsAdminOfPredeploysForChain(t, *chain)
		})
	}
}
