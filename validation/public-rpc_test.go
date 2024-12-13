package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func testPublicRPC(t *testing.T, chain *ChainConfig) {
	rpcEndpoint := chain.PublicRPC
	require.NotEmpty(t, rpcEndpoint, "no public_rpc endpoint specified")

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint '%s'", rpcEndpoint)
	defer client.Close()
}
