package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func testKeyHandover(t *testing.T, chain *ChainConfig) {
	chainID := chain.ChainID
	superchain := OPChains[chainID].Superchain
	rpcEndpoint := Superchains[superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint, "no rpc specified")

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	// L1 Proxy Admin
	checkResolutions(t, standard.Config.MultisigRoles[superchain].KeyHandover.L1.Universal, chainID, client)

	rpcEndpoint = OPChains[chainID].PublicRPC
	require.NotEmpty(t, rpcEndpoint, "no rpc specified")

	client, err = ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	// L2 Proxy Admin
	checkResolutions(t, standard.Config.MultisigRoles[superchain].KeyHandover.L2.Universal, chainID, client)
}
