package validation

import (
	"context"
	"testing"

	"github.com/ethereum-optimism/optimism/op-service/retry"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum"
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

	contractCallResolutions := []struct {
		name                     string
		method                   string
		shouldResolveToAddressOf string
	}{
		{"AddressManager", "owner()", "ProxyAdmin"},
		{"SystemConfigProxy", "owner()", "SystemConfigOwner"},
		// {"DisputeGameFactoryProxy", "owner", "ProxyAdminOwner"}, // TODO reinstate this but only run the check if the chain is on FPAC or greater
		// {"DelayedWETHProxy", "owner", "ProxyAdminOwner"},        // TODO reinstate this but only run the check if the chain is on FPAC or greater
		{"ProxyAdmin", "owner()", "ProxyAdminOwner"},
	}

	for _, r := range contractCallResolutions {
		contractAddress, err := Addresses[chainID].AddressFor(r.name)
		require.NoError(t, err)

		want, err := Addresses[chainID].AddressFor(r.shouldResolveToAddressOf)
		require.NoError(t, err)

		got, err := getAddressWithRetries(r.method, common.Address(contractAddress), client)
		require.NoError(t, err)

		assert.Equal(t, want, got, "%s.%s = %s, expected %s (%s)", r.name, r.method, want, r.shouldResolveToAddressOf)
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
	const maxAttempts = 10
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
