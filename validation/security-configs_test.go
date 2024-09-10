package validation

import (
	"context"
	"strconv"
	"strings"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/internal/bindings"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

var checkResolutions = func(t *testing.T, r standard.Resolutions, chainID uint64, client *ethclient.Client) {
	for contract, methodToOutput := range r {

		var contractAddress Address
		var err error

		if common.IsHexAddress(contract) {
			contractAddress = Address(common.HexToAddress(contract))
		} else {
			contractAddress, err = Addresses[chainID].AddressFor(contract)
			require.NoError(t, err)
		}

		for method, output := range methodToOutput {

			var want Address

			if common.IsHexAddress(output) {
				want = Address(common.HexToAddress(output))
			} else {
				want, err = Addresses[chainID].AddressFor(output)
				require.NoError(t, err)
			}

			got, err := getAddress(method, contractAddress, client)
			require.NoErrorf(t, err, "problem calling %s.%s (%s)", contract, method, contractAddress)

			// Use t.Errorf here for a concise output of failures, since failure info is sent to a slack channel
			if want != got {
				t.Errorf("%s.%s = %s, expected %s (%s)", contract, method, got, want, output)
			}
		}

	}
}

func testL1SecurityConfig(t *testing.T, chain *ChainConfig) {
	chainID := chain.ChainID

	rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC
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
	isFaultProofs := majorVersion >= 3

	checkResolutions(t, standard.Config.Roles.L1.GetResolutions(isFaultProofs), chainID, client)

	checkResolutions(t, standard.Config.MultisigRoles[OPChains[chainID].Superchain].L1.GetResolutions(isFaultProofs), chainID, client)

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
	actualAddress := Address(common.BytesToAddress(actualAddressManagerBytes[12:32]))
	if am != actualAddress {
		t.Errorf("AddressManager should be: %s, got: %s", am, actualAddress)
	}
}

func testL2SecurityConfig(t *testing.T, chain *ChainConfig) {
	// Create an ethclient connection to the specified RPC URL
	client, err := ethclient.Dial(chain.PublicRPC)
	require.NoError(t, err, "Failed to connect to the Ethereum client at RPC url %s", chain.PublicRPC)
	defer client.Close()
	checkResolutions(t, standard.Config.Roles.L2.Universal, chain.ChainID, client)
}

func getAddress(method string, contractAddress Address, client *ethclient.Client) (Address, error) {
	addr := (common.Address(contractAddress))
	callMsg := ethereum.CallMsg{
		To:   &addr,
		Data: crypto.Keccak256([]byte(method))[:4],
	}

	// Make the call
	callContract := func(msg ethereum.CallMsg) ([]byte, error) {
		return client.CallContract(context.Background(), msg, nil)
	}
	result, err := Retry(callContract)(callMsg)
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
