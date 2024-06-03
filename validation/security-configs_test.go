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

	addressManagerAddress, err := Addresses[chainID].AddressFor("AddressManager")
	require.NoError(t, err)

	proxyAdmin, err := Addresses[chainID].AddressFor("ProxyAdmin")
	require.NoError(t, err)

	owner, err := getOwnerWithRetries(common.Address(addressManagerAddress), client)
	require.NoError(t, err)

	assert.Equal(t, proxyAdmin, owner, "AddressManager.Owner() = %s, expected %s (ProxyAdmin)", owner, proxyAdmin)

}

func TestSecurityConfigs(t *testing.T) {
	isExcluded := map[uint64]bool{}
	for chainID, chain := range OPChains {
		t.Run(perChainTestName(chain), func(t *testing.T) {
			if isExcluded[chain.ChainID] {
				t.Skipf("chain %d: EXCLUDED from Genesis block hash validation", chainID)
			}
			testSecurityConfigOfChain(t, chainID)
		})
	}
}

// getOwnerWithRetries is a wrapper for getOwner
// which retries up to 10 times with exponential backoff.
func getOwnerWithRetries(addr common.Address, client *ethclient.Client) (Address, error) {
	const maxAttempts = 10
	return retry.Do(context.Background(), maxAttempts, retry.Exponential(), func() (Address, error) {
		return getOwner(addr, client)
	})
}

func getOwner(contractAddress common.Address, client *ethclient.Client) (Address, error) {

	callMsg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: crypto.Keccak256([]byte("owner()"))[:4],
	}

	// Make the call
	result, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return Address{}, err
	}

	return Address(common.BytesToAddress(result)), nil
}
