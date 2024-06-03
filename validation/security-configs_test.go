package validation

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-service/retry"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSecurityConfigOfChain(t *testing.T, chainID uint64) {

	rpcEndpoint := Superchains[OPChains[chainID].Superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint, "no rpc specified")

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	type resolution struct {
		name                     string
		method                   string
		shouldResolveToAddressOf string
	}

	contractCallResolutions := []resolution{
		{"AddressManager", "owner()", "ProxyAdmin"},
		{"SystemConfigProxy", "owner()", "SystemConfigOwner"},
		{"ProxyAdmin", "owner()", "ProxyAdminOwner"},
		{"L1CrossDomainMessengerProxy", "PORTAL()", "OptimismPortalProxy"},
		{"L1ERC721BridgeProxy", "admin()", "ProxyAdmin"},
		{"L1ERC721BridgeProxy", "messenger()", "L1CrossDomainMessengerProxy"},
		{"L1StandardBridgeProxy", "getOwner()", "ProxyAdmin"},
		{"L1StandardBridgeProxy", "messenger()", "L1CrossDomainMessengerProxy"},
		{"OptimismMintableERC20FactoryProxy", "admin()", "ProxyAdmin"},
		{"OptimismMintableERC20FactoryProxy", "BRIDGE()", "L1StandardBridgeProxy"},
		{"OptimismPortalProxy", "admin()", "ProxyAdmin"},
		{"ProxyAdmin", "owner()", "ProxyAdminOwner"},
		{"ProxyAdmin", "addressManager()", "AddressManager"},
		{"SystemConfigProxy", "admin()", "ProxyAdmin"},
		{"SystemConfigProxy", "owner()", "SystemConfigOwner"},
	}

	portalProxyAddress, err := Addresses[chainID].AddressFor("OptimismPortalProxy")
	require.NoError(t, err)
	portalProxy, err := bindings.NewOptimismPortal(common.Address(portalProxyAddress), client)
	require.NoError(t, err)
	version, err := portalProxy.Version(&bind.CallOpts{})
	require.NoError(t, err)
	majorVersion, err := strconv.ParseInt(strings.Split(version, ".")[0], 10, 32)
	require.NoError(t, err)

	// Portal version `3` is the first version of the `OptimismPortal` that supported the fault proof system.
	isFPAC := majorVersion >= 3

	if isFPAC {
		contractCallResolutions = append(contractCallResolutions,
			resolution{"DisputeGameFactoryProxy", "admin()", "ProxyAdmin"},
			resolution{"AnchorStateRegistryProxy", "admin()", "ProxyAdmin"},
			resolution{"DelayedWETHProxy", "admin()", "ProxyAdmin"},
			resolution{"DelayedWETHProxy", "admin()", "ProxyAdmin"},
			resolution{"DelayedWETHProxy", "owner()", "ProxyAdminOwner"},
			resolution{"OptimismPortalProxy", "guardian()", "Guardian"},
			resolution{"OptimismPortalProxy", "systemConfig()", "SystemConfigProxy"},
		)
	} else {
		contractCallResolutions = append(contractCallResolutions,
			resolution{"OptimismPortalProxy", "GUARDIAN()", "Guardian"},
			resolution{"OptimismPortalProxy", "SYSTEM_CONFIG()", "SystemConfigProxy"},
			resolution{"OptimismPortalProxy", "L2_ORACLE()", "L2OutputOracleProxy"},
			resolution{"L2OutputOracleProxy", "admin()", "ProxyAdmin"},
			resolution{"L2OutputOracleProxy", "CHALLENGER()", "Challenger"},
		)
	}

	for _, r := range contractCallResolutions {
		contractAddress, err := Addresses[chainID].AddressFor(r.name)
		require.NoError(t, err)

		want, err := Addresses[chainID].AddressFor(r.shouldResolveToAddressOf)
		require.NoError(t, err)

		got, err := getAddressWithRetries(r.method, common.Address(contractAddress), client)
		require.NoErrorf(t, err, "problem calling %s.%s", contractAddress, r.method)

		assert.Equal(t, want, got, "%s.%s = %s, expected %s (%s)", r.name, r.method, got, want, r.shouldResolveToAddressOf)
	}

}

func TestSecurityConfigs(t *testing.T) {
	isExcluded := map[uint64]bool{
		11155421: true, // OP_Labs_Sepolia_devnet_0  (no SystemConfigOwner specified)
		11763072: true, // Base_devnet_0 (no SystemConfigOwner specified)
	}
	for chainID, chain := range OPChains {
		t.Run(perChainTestName(chain), func(t *testing.T) {
			if isExcluded[chain.ChainID] {
				t.Skipf("chain %d: EXCLUDED from Security Config Checks", chainID)
			}
			testSecurityConfigOfChain(t, chainID)
		})
	}
}

// getAddressWithRetries is a wrapper for getAddress
// which retries up to 10 times with exponential backoff.
func getAddressWithRetries(method string, addr common.Address, client *ethclient.Client) (Address, error) {
	const maxAttempts = 1
	return retry.Do(context.Background(), maxAttempts, retry.Exponential(), func() (Address, error) {
		return getAddress(method, addr, client)
	})
}

func getAddress(method string, contractAddress common.Address, client *ethclient.Client) (Address, error) {

	callMsg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: crypto.Keccak256([]byte(method))[:4],
	}

	// Make the call
	result, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return Address{}, err
	}

	return Address(common.BytesToAddress(result)), nil
}
