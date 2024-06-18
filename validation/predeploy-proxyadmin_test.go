package validation

import (
	"strings"
	"testing"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

	proxyAdminAddr := predeploys.ProxyAdminAddr

	proxyAdmin, err := bindings.NewProxyAdmin(proxyAdminAddr, client)
	require.NoError(t, err)

	getProxyAdmin := func(a common.Address) (common.Address, error) {
		return proxyAdmin.GetProxyAdmin(&bind.CallOpts{}, a)
	}

	for k, p := range predeploys.Predeploys {
		t.Run(k, func(t *testing.T) {
			if !strings.HasPrefix(p.Address.Hex(), "0x42") {
				t.Skipf("%s is not a predeploy (probably a preinstall)", k)
			}
			if k == "GovernanceToken" || k == "WETH" || k == "WETH9" {
				t.Skipf("%s excluded", k)
			}
			proxyAdminOf, err := Retry(getProxyAdmin)(p.Address)
			assert.NoError(t, err)
			assert.Equal(t, proxyAdminOf, proxyAdminAddr, "ProxyAdmin.getProxyAdmin(%s) = %s, expected %s", p.Address, proxyAdminOf, proxyAdminAddr)
		})
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
