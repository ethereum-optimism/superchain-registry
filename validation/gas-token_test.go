package validation

import (
	"context"
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
	// Create an ethclient connection to the specified RPC URL
	client, err := ethclient.Dial(chain.PublicRPC)
	require.NoError(t, err, "Failed to connect to the Ethereum client at RPC url %s", chain.PublicRPC)
	defer client.Close()

	want := "0000000000000000000000000000000000000000000000000000000000000020" + // offset
		"000000000000000000000000000000000000000000000000000000000000000d" + // length
		"5772617070656420457468657200000000000000000000000000000000000000" // "Wrapped Ether" padded to 32 bytes

	got, err := getString("name()", Address(common.HexToAddress("0x4200000000000000000000000000000000000006")), client)
	require.NoError(t, err)

	require.Equal(t, want, got)
}

func getString(method string, contractAddress Address, client *ethclient.Client) (interface{}, error) {
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
