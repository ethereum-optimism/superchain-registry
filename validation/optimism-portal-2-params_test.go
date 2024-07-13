package validation

import (
	"context"
	"testing"

	bindings "github.com/ethereum-optimism/optimism/op-node/bindings/preview"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testOptimismPortal2Params(t *testing.T, chain *ChainConfig) {
	opAddr, err := Addresses[chain.ChainID].AddressFor("OptimismPortalProxy")
	require.NoError(t, err)
	t.Log(opAddr)

	rpcEndpoint := chain.PublicRPC
	require.NotEmpty(t, rpcEndpoint, "no public endpoint for chain")
	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	op, err := bindings.NewOptimismPortal2(common.Address(opAddr), client)
	require.NoError(t, err)

	callOpts := &bind.CallOpts{Context: context.Background()}

	std := standard.Config.Params[chain.Superchain].OptimismPortal2Config

	pmds, err := op.ProofMaturityDelaySeconds(callOpts)
	require.NoError(t, err)
	assertIntInBounds(t, "Proof Maturity Delay", pmds.Uint64(), std.ProofMaturityDelaySeconds)

	dgfds, err := op.DisputeGameFinalityDelaySeconds(callOpts)
	require.NoError(t, err)
	assertIntInBounds(t, "Proof Maturity Delay", dgfds.Uint64(), std.DisputeGameFinalityDelaySeconds)

	rgt, err := op.RespectedGameType(callOpts)
	require.NoError(t, err)
	assert.Equal(t, rgt, std.RespectedGameType)
}
