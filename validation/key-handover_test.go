package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var Multisig = map[string]Address{
	// On each superchain, this is a 2/2 Safe between the Optimism Foundation and the Security Council
	"mainnet":       MustHexToAddress("0x5a0Aae59D09fccBdDb6C6CcEB07B7279367C3d2A"),
	"sepolia":       MustHexToAddress("0x1Eb2fFc903729a0F03966B917003800b145F56E2"),
	"sepolia-dev-0": MustHexToAddress("0x4377BB0F0103992b31eC12b4d796a8687B8dC8E9"),
}

func testKeyHandoverOfChain(t *testing.T, chainID uint64) {
	superchain := OPChains[chainID].Superchain
	rpcEndpoint := Superchains[superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint, "no rpc specified")

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	proxyAdmin, err := Addresses[chainID].AddressFor("ProxyAdmin")
	require.NoError(t, err)

	want := Multisig[superchain]
	got, err := getAddress("owner()", proxyAdmin, client)
	require.NoError(t, err)

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
