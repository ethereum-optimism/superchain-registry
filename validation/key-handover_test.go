package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var L1Multisig = map[string]Address{
	// On each superchain, this is a 2/2 Safe between the Optimism Foundation and the Security Council
	"mainnet":       MustHexToAddress("0x5a0Aae59D09fccBdDb6C6CcEB07B7279367C3d2A"),
	"sepolia":       MustHexToAddress("0x1Eb2fFc903729a0F03966B917003800b145F56E2"),
	"sepolia-dev-0": MustHexToAddress("0x4377BB0F0103992b31eC12b4d796a8687B8dC8E9"),
}

var L2Multisig = map[string]Address{
	// On each superchain, this is the ALIASED address of
	// the L1 2/2 Safe between the Optimism Foundation and the Security Council.
	// To compute the aliased address, add 0x1111000000000000000000000000000000001111
	"mainnet":       MustHexToAddress("0x6B1BAE59D09fCcbdDB6C6cceb07B7279367C4E3b"),
	"sepolia":       MustHexToAddress("0x2FC3ffc903729a0f03966b917003800B145F67F3"),
	"sepolia-dev-0": MustHexToAddress("0x5488bb0f0103992b31eC12B4D796A8687B8Dd9fa"),
}

func testKeyHandoverOfChain(t *testing.T, chainID uint64) {
	superchain := OPChains[chainID].Superchain
	rpcEndpoint := Superchains[superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint, "no rpc specified")

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	// L1 Proxy Admin
	proxyAdmin, err := Addresses[chainID].AddressFor("ProxyAdmin")
	require.NoError(t, err)

	want := L1Multisig[superchain]
	got, err := getAddress("owner()", proxyAdmin, client)
	require.NoError(t, err)

	assert.Equal(t, want, got, "L1: %s.%s = %s, expected %s (%s)", proxyAdmin, "owner()", got, want, "ProxyAdminOwner")

	// L2 Proxy Admin
	proxyAdmin = MustHexToAddress("0x4200000000000000000000000000000000000018")
	client, err = ethclient.Dial(OPChains[chainID].PublicRPC)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", OPChains[chainID].PublicRPC)

	want = L2Multisig[superchain]
	got, err = getAddress("owner()", proxyAdmin, client)
	require.NoError(t, err)

	assert.Equal(t, want, got, "L2: %s.%s = %s, expected %s (%s)", proxyAdmin, "owner()", got, want, "ProxyAdminOwner")
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
