package validation

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-service/retry"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
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

	contractCallResolutions := standard.Config[OPChains[chainID].Superchain].L1.GetResolutions(isFPAC)

	for contract, methodToOutput := range contractCallResolutions {

		contractAddress, err := Addresses[chainID].AddressFor(contract)
		require.NoError(t, err)

		for method, output := range methodToOutput {

			var want Address
			if strings.HasPrefix(output, "0x") {
				want = MustHexToAddress(output)
			} else {
				want, err = Addresses[chainID].AddressFor(output)
				require.NoError(t, err)
			}

			got, err := getAddressWithRetries(method, contractAddress, client)
			require.NoErrorf(t, err, "problem calling %s.%s %s", contract, contractAddress, method)

			assert.Equal(t, want, got, "%s.%s = %s, expected %s (%s)", contract, method, got, want, output)
		}

	}

	// Perform an extra check on a mapping value of "L1CrossDomainMessengerProxy":
	// This is because L1CrossDomainMessenger's proxy is a ResolvedDelegateProxy, and
	// ResolvedDelegateProxy does not expose a getter method to tell us who can update its implementations (i.e. similar to the proxy admin role for a regular proxy).
	// So to ensure L1CrossDomainMessengerProxy's "proxy admin" is properly set up, we need to peek into L1CrossDomainMessengerProxy(aka ResolvedDelegateProxy)'s storage
	// slot to get the value of addressManager[address(this)], and ensure it is the expected AddressManager address, and together with the owner check within AddressManager,
	// we now have assurance that L1CrossDomainMessenger's proxy is properly managed.
	l1cdmp, err := Addresses[chainID].AddressFor("L1CrossDomainMessengerProxy")
	require.NoError(t, err)
	actualAddressManagerBytes, err := getMappingValue(l1cdmp, 1, l1cdmp, client)
	require.NoError(t, err)
	am, err := Addresses[chainID].AddressFor("AddressManager")
	require.NoError(t, err)
	assert.Equal(t,
		am[:],
		actualAddressManagerBytes[12:32],
	)
}

func TestSecurityConfigs(t *testing.T) {
	isExcluded := map[uint64]bool{
		11155421: true, // OP_Labs_Sepolia_devnet_0  (no AnchorStateRegistryProxy specified)
		11763072: true, // Base_devnet_0 (no AnchorStateRegistryProxy specified)
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
func getAddressWithRetries(method string, addr Address, client *ethclient.Client) (Address, error) {
	const maxAttempts = 10
	return retry.Do(context.Background(), maxAttempts, retry.Exponential(), func() (Address, error) {
		return getAddress(method, addr, client)
	})
}

func getAddress(method string, contractAddress Address, client *ethclient.Client) (Address, error) {
	addr := (common.Address(contractAddress))
	callMsg := ethereum.CallMsg{
		To:   &addr,
		Data: crypto.Keccak256([]byte(method))[:4],
	}

	// Make the call
	result, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return Address{}, err
	}

	return Address(common.BytesToAddress(result)), nil
}

func getMappingValue(contractAddress Address, mapSlot uint8, key Address, client *ethclient.Client) ([]byte, error) {
	preimage := make([]byte, 12, 64)
	preimage = append(preimage, key[:]...)
	pad := [31]byte{}
	preimage = append(preimage, pad[:]...)
	preimage = append(preimage, mapSlot)
	storageSlot := crypto.Keccak256Hash(preimage)
	return client.StorageAt(context.Background(), common.Address(contractAddress), storageSlot, nil)
}
