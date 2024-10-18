package validation

import (
	"context"
	"strings"
	"testing"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func testFaultGameParams(t *testing.T, chain *ChainConfig) {
	rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint, "no public endpoint for L1 chain")

	clientL1, err := ethclient.Dial(rpcEndpoint)
	require.NoError(t, err, "Failed to connect to the Ethereum client at RPC url %s", rpcEndpoint)
	defer clientL1.Close()

	clientL2, err := ethclient.Dial(chain.PublicRPC)
	require.NoError(t, err, "Failed to connect to the l2 client at RPC url %s", chain.PublicRPC)
	defer clientL2.Close()

	permissionedDisputeGameAddr, err := Addresses[chain.ChainID].AddressFor("PermissionedDisputeGame")
	require.NoError(t, err)

	preimageOracleAddr, err := Addresses[chain.ChainID].AddressFor("PreimageOracle")
	require.NoError(t, err)

	delayedWethAddr, err := Addresses[chain.ChainID].AddressFor("DelayedWETHProxy")
	require.NoError(t, err)

	anchorStateRegistryAddr, err := Addresses[chain.ChainID].AddressFor("AnchorStateRegistryProxy")
	require.NoError(t, err)

	optimismPortalAddr, err := Addresses[chain.ChainID].AddressFor("OptimismPortalProxy")
	require.NoError(t, err)

	// OptimismPortal: check for permissioned vs permissionless game type
	respectedGameType, err := CastCall(optimismPortalAddr, "respectedGameType()", nil, rpcEndpoint)
	require.NoError(t, err)

	var isPermissionless bool
	switch respectedGameType[0] {
	case "0x0000000000000000000000000000000000000000000000000000000000000000":
		isPermissionless = true
		t.Log("detected Permissionless game type")
	case "0x0000000000000000000000000000000000000000000000000000000000000001":
		isPermissionless = false
		t.Log("detected Permissioned game type")
	default:
		require.Fail(t, "unexpected return value from OptimismPortalProxy.respectedGameType()")
	}

	// PermissionedDisputeGame
	maxGameDepth, err := CastCall(permissionedDisputeGameAddr, "maxGameDepth()", nil, rpcEndpoint)
	require.NoError(t, err)
	require.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000049", maxGameDepth[0], "PermissionedDisputeGame: fault game max depth") // 73

	splitDepth, err := CastCall(permissionedDisputeGameAddr, "splitDepth()", nil, rpcEndpoint)
	require.NoError(t, err)
	require.Equal(t, "0x000000000000000000000000000000000000000000000000000000000000001e", splitDepth[0], "PermissionedDisputeGame: fault game split depth") // 30

	maxClockDuration, err := CastCall(permissionedDisputeGameAddr, "maxClockDuration()", nil, rpcEndpoint)
	require.NoError(t, err)
	require.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000049d40", maxClockDuration[0], "PermissionedDisputeGame: max game clock duration") // 302400 sec = 3.5 days

	clockExtension, err := CastCall(permissionedDisputeGameAddr, "clockExtension()", nil, rpcEndpoint)
	require.NoError(t, err)
	require.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000002a30", clockExtension[0], "PermissionedDisputeGame: game clock extension") // 10800 sec = 3 hours

	absolutePrestate, err := CastCall(permissionedDisputeGameAddr, "absolutePrestate()", nil, rpcEndpoint)
	require.NoError(t, err)
	require.Truef(t, findOpProgramRelease(absolutePrestate[0]), "onchain op-program prestate hash is not from a standard version: %v", absolutePrestate[0])

	l2BlockNumber, err := CastCall(permissionedDisputeGameAddr, "l2BlockNumber()", nil, rpcEndpoint)
	require.NoError(t, err)
	// 0 for chains using fault proofs from genesis as per spec https://specs.optimism.io/protocol/configurability.html#fault-game-genesis-block
	require.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000000", l2BlockNumber[0], "PermissionedDisputeGame: fault game genesis block")

	// PreimageOracle
	challengePeriod, err := CastCall(preimageOracleAddr, "challengePeriod()", nil, rpcEndpoint)
	require.NoError(t, err)
	require.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000015180", challengePeriod[0], "PreimageOracle: large preimage proposal challenge period") // 86400 sec = 24 hours

	minProposalSize, err := CastCall(preimageOracleAddr, "minProposalSize()", nil, rpcEndpoint)
	require.NoError(t, err)
	require.Equal(t, "0x000000000000000000000000000000000000000000000000000000000001ec30", minProposalSize[0], "PreimageOracle: minimum large preimage proposal size") // 126000 bytes

	// DelayedWETH
	wethDelay, err := CastCall(delayedWethAddr, "delay()", nil, rpcEndpoint)
	require.NoError(t, err)
	require.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000093a80", wethDelay[0], "DelayedWETH: bond withdrawal delay") // 604800 sec = 7 days

	// AnchorStateRegistry
	var anchors []string
	if isPermissionless {
		anchors, err = CastCall(anchorStateRegistryAddr, "anchors(uint32)(bytes32,uint256)", []string{"0"}, rpcEndpoint)
	} else {
		anchors, err = CastCall(anchorStateRegistryAddr, "anchors(uint32)(bytes32,uint256)", []string{"1"}, rpcEndpoint)
	}
	require.NoError(t, err)

	var out *eth.OutputResponse
	err = clientL2.Client().CallContext(context.Background(), &out, "optimism_outputAtBlock", anchors[1])
	require.NoError(t, err)
	require.Equal(t, out.OutputRoot.String(), anchors[0], "AnchorStateRegistry: output root hash")
}

func findOpProgramRelease(hash string) bool {
	for _, element := range standard.OpProgramReleases.Releases {
		if strings.EqualFold(element.Hash, hash) {
			return true
		}
	}
	return false
}
