package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testKeyHandoverOfChain(t *testing.T, chainID uint64) {
	rpcEndpoint := Superchains[OPChains[chainID].Superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint, "no rpc specified")

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	proxyAdmin, err := Addresses[chainID].AddressFor("ProxyAdmin")
	require.NoError(t, err)

	got, err := getAddressWithRetries("owner()", proxyAdmin, client)
	require.NoError(t, err)

	want := Address(common.HexToAddress("0x5a0Aae59D09fccBdDb6C6CcEB07B7279367C3d2A"))

	assert.Equal(t, want, got, "%s.%s = %s, expected %s (%s)", proxyAdmin, "owner()", got, want, "ProxyAdminOwner")
}

func TestKeyHandover(t *testing.T) {
	for chainID, chain := range OPChains {
		t.Run(perChainTestName(chain), func(t *testing.T) {
			RunOnlyOnStandardChains(t, *chain)
			testKeyHandoverOfChain(t, chainID)
		})
	}
}
