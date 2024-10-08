package validation

import (
	"context"
	"errors"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func testGasToken(t *testing.T, chain *ChainConfig) {
	client, err := ethclient.Dial(chain.PublicRPC)
	require.NoError(t, err, "Failed to connect to the Ethereum client at RPC url %s", chain.PublicRPC)
	defer client.Close()

	// WETH predeploy name() check
	want := "0000000000000000000000000000000000000000000000000000000000000020" + // offset
		"000000000000000000000000000000000000000000000000000000000000000d" + // length
		"5772617070656420457468657200000000000000000000000000000000000000" // "Wrapped Ether" padded to 32 bytes
	gotName, err := getHexString("name()", MustHexToAddress("0x4200000000000000000000000000000000000006"), client)
	require.NoError(t, err)
	require.Equal(t, want, gotName)

	// L1Block .isCustomGasToken() check
	got, err := getBool("isCustomGasToken()", MustHexToAddress("0x4200000000000000000000000000000000000015"), client)

	if err != nil {
		// Pre: Custom Gas Token feature. Reverting is acceptable.
		require.Contains(t, err.Error(), "execution reverted")
	} else {
		// Post: Custom Gas Token fearure. Must be set to false.
		require.False(t, got)
	}

	// SystemConfig .isCustomGasToken() check (L1)
	client, err = ethclient.Dial(Superchains[chain.Superchain].Config.L1.PublicRPC)
	require.NoError(t, err, "Failed to connect to the Ethereum client at RPC url %s", chain.PublicRPC)
	defer client.Close()

	got, err = getBool("isCustomGasToken()", Addresses[chain.ChainID].SystemConfigProxy, client)
	if err != nil {
		// Pre: Custom Gas Token feature. Reverting is acceptable.
		require.Contains(t, err.Error(), "execution reverted")
	} else {
		// Post: Custom Gas Token fearure. Must be set to false.
		require.False(t, got)
	}
}

func getBytes(method string, contractAddress Address, client *ethclient.Client) ([]byte, error) {
	addr := (common.Address(contractAddress))
	callMsg := ethereum.CallMsg{
		To:   &addr,
		Data: crypto.Keccak256([]byte(method))[:4],
	}

	// Make the call
	callContract := func(msg ethereum.CallMsg) ([]byte, error) {
		return client.CallContract(context.Background(), msg, nil)
	}

	return Retry(callContract)(callMsg)
}

func getHexString(method string, contractAddress Address, client *ethclient.Client) (string, error) {
	result, err := getBytes(method, contractAddress, client)

	return common.Bytes2Hex(result), err
}

func getBool(method string, contractAddress Address, client *ethclient.Client) (bool, error) {
	result, err := getBytes(method, contractAddress, client)
	if err != nil {
		return false, err
	}

	switch common.HexToHash(string(result)) {
	case common.Hash{1}:
		return true, nil
	case common.Hash{}:
		return false, nil
	default:
		return false, errors.New("unexpected non-bool return value")
	}
}
