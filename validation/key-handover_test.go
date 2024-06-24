package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func testKeyHandoverOfChain(t *testing.T, chainID uint64) {
	superchain := OPChains[chainID].Superchain
	rpcEndpoint := Superchains[superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint, "no rpc specified")

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	// L1 Proxy Admin
	checkResolutions(t, standard.Config.MultisigRoles[superchain].KeyHandover.L1.Universal, chainID, client)

	// L2 Proxy Admin
	checkResolutions(t, standard.Config.MultisigRoles[superchain].KeyHandover.L1.Universal, chainID, client)
}

func TestKeyHandover(t *testing.T) {
	isExcluded := map[uint64]bool{
		11155421: true, // OP Labs Sepolia devnet 0 (no rpc endpoint)
	}
	for chainID, chain := range OPChains {
		t.Run(perChainTestName(chain), func(t *testing.T) {
			if isExcluded[chainID] {
				t.Skip("chain excluded from key handover check")
			}
			RunOnlyOnStandardChains(t, *chain)
			testKeyHandoverOfChain(t, chainID)
		})
	}
}
