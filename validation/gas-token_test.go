package validation

import (
	"context"
	"errors"
	"strings"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func testGasToken(t *testing.T, chain *ChainConfig) {
	skipIfExcluded(t, chain.ChainID)

	client, err := ethclient.Dial(chain.PublicRPC)
	require.NoError(t, err, "Failed to connect to the Ethereum client at RPC url %s", chain.PublicRPC)
	defer client.Close()

	// WETH predeploy name() check
	want := "0000000000000000000000000000000000000000000000000000000000000020" + // offset
		"000000000000000000000000000000000000000000000000000000000000000d" + // length
		"5772617070656420457468657200000000000000000000000000000000000000" // "Wrapped Ether" padded to 32 bytes
	gotName, err := getString("name()", MustHexToAddress("0x4200000000000000000000000000000000000006"), client)
	require.NoError(t, err)
	require.Equal(t, want, gotName)

	// L1Block .isCustomGasToken() check
	got, err := getBool("isCustomGasToken()", MustHexToAddress("0x4200000000000000000000000000000000000015"), client)
	if !strings.Contains(err.Error(), "execution reverted") {
		// Pre: Custom Gas Token feature. Reverting is acceptable.
		require.NoError(t, err)
	} else {
		// Post: Custom Gas Token fearure. Must be set to false.
		require.False(t, got)
	}

	// SystemConfig .isCustomGasToken() check (L1)
	client, err = ethclient.Dial(Superchains[chain.Superchain].Config.L1.PublicRPC)
	require.NoError(t, err, "Failed to connect to the Ethereum client at RPC url %s", chain.PublicRPC)
	defer client.Close()

	got, err = getBool("isCustomGasToken()", Addresses[chain.ChainID].SystemConfigProxy, client)
	if !strings.Contains(err.Error(), "execution reverted") {
		// Pre: Custom Gas Token feature. Reverting is acceptable.
		require.NoError(t, err)
	} else {
		// Post: Custom Gas Token fearure. Must be set to false.
		require.False(t, got)
	}
}

func getString(method string, contractAddress Address, client *ethclient.Client) (string, error) {
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

	return common.Bytes2Hex(result), err
}

func getBool(method string, contractAddress Address, client *ethclient.Client) (bool, error) {
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
		return false, err
	}

	// Decode the result
	var decodedBool bool

	// If the result is a string "0x1" or "0x0", convert to bool
	if string(result) == "0x1" {
		decodedBool = true
	} else if string(result) == "0x0" {
		decodedBool = false
	} else {
		return false, errors.New("Unexpected return value, cannot convert to bool")
	}

	return decodedBool, nil
}
