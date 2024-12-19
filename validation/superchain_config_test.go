package validation

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/testutils"
)

var (
	testChainID            = uint64(10)
	delayedWETHProxyString = "0xabcdef0000000000000000000000000000000000"
	normalContractString   = "0x1234560000000000000000000000000000000000"
	expectedAddressString  = "0xaaaaaa0000000000000000000000000000000000"
)

// Setup some helper addresses and methods
func TestVerifyOnchain_Success(t *testing.T) {
	t.Parallel()
	delayedWETHProxy := superchain.MustHexToAddress(delayedWETHProxyString)
	normalContract := superchain.MustHexToAddress(normalContractString)
	expected := superchain.MustHexToAddress(expectedAddressString)
	targetContracts := []superchain.Address{delayedWETHProxy, normalContract}

	// For DelayedWETHProxy: method = "config()"
	configMethodID := string(testutils.MethodID("config()"))
	configResponse := common.HexToAddress(expectedAddressString).Bytes()

	// For other contracts: method = "superchainConfig()"
	superchainConfigMethodID := string(testutils.MethodID("superchainConfig()"))
	superchainConfigResponse := common.HexToAddress(expectedAddressString).Bytes()

	mockClient := &testutils.MockEthClient{
		Responses: map[string][]byte{
			configMethodID:           configResponse,
			superchainConfigMethodID: superchainConfigResponse,
		},
	}

	err := verifySuperchainConfigOnchain(testChainID, mockClient, targetContracts, expected)
	require.NoError(t, err)
}

func TestVerifyOnchain_ErrorFromCall(t *testing.T) {
	t.Parallel()
	delayedWETHProxy := superchain.MustHexToAddress(delayedWETHProxyString)
	expected := superchain.MustHexToAddress("0xabcdef0000000000000000000000000000000000")
	targetContracts := []superchain.Address{delayedWETHProxy}

	mockClient := &testutils.MockEthClient{
		Err: errors.New("test call contract error"),
	}

	err := verifySuperchainConfigOnchain(testChainID, mockClient, targetContracts, expected)
	require.Error(t, err)
	require.Contains(t, err.Error(), "test call contract error")
}

func TestVerifyOnchain_IncorrectAddress(t *testing.T) {
	t.Parallel()
	delayedWETHProxy := superchain.MustHexToAddress(delayedWETHProxyString)
	normalContract := superchain.MustHexToAddress(normalContractString)
	expected := superchain.MustHexToAddress(expectedAddressString)
	targetContracts := []superchain.Address{delayedWETHProxy, normalContract}

	// Response returns a different address than expected
	mockClient := &testutils.MockEthClient{
		Responses: map[string][]byte{
			string(testutils.MethodID("superchainConfig()")): common.HexToAddress("0x9999990000000000000000000000000000000000").Bytes(),
		},
	}

	err := verifySuperchainConfigOnchain(testChainID, mockClient, targetContracts, expected)
	require.Error(t, err)
	require.Contains(t, err.Error(), "incorrect superchainConfig() address")
	require.Contains(t, err.Error(), "wanted 0xAAaAaA0000000000000000000000000000000000")
	require.Contains(t, err.Error(), "got 0x9999990000000000000000000000000000000000")
}
