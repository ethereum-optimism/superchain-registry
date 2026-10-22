package report

import (
	"testing"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func testV7DeployOutput() DeployOPChainOutput {
	return DeployOPChainOutput{
		OpChainProxyAdmin:                  common.BigToAddress(common.Big1),
		AddressManager:                     common.BigToAddress(common.Big2),
		L1ERC721BridgeProxy:                common.HexToAddress("0x3"),
		SystemConfigProxy:                  common.HexToAddress("0x4"),
		OptimismMintableERC20FactoryProxy:  common.HexToAddress("0x5"),
		L1StandardBridgeProxy:              common.HexToAddress("0x6"),
		L1CrossDomainMessengerProxy:        common.HexToAddress("0x7"),
		EthLockboxProxy:                    common.HexToAddress("0x8"),
		OptimismPortalProxy:                common.HexToAddress("0x9"),
		DisputeGameFactoryProxy:            common.HexToAddress("0xa"),
		AnchorStateRegistryProxy:           common.HexToAddress("0xb"),
		FaultDisputeGame:                   common.HexToAddress("0xc"),
		PermissionedDisputeGame:            common.HexToAddress("0xd"),
		DelayedWETHPermissionedGameProxy:   common.HexToAddress("0xe"),
		DelayedWETHPermissionlessGameProxy: common.HexToAddress("0xf"),
	}
}

func TestDeployOutputFromStateV7(t *testing.T) {
	expected := testV7DeployOutput()
	deployment := map[string]any{
		"OpChainProxyAdminImpl":              expected.OpChainProxyAdmin.Hex(),
		"AddressManagerImpl":                 expected.AddressManager.Hex(),
		"L1Erc721BridgeProxy":                expected.L1ERC721BridgeProxy.Hex(),
		"SystemConfigProxy":                  expected.SystemConfigProxy.Hex(),
		"OptimismMintableErc20FactoryProxy":  expected.OptimismMintableERC20FactoryProxy.Hex(),
		"L1StandardBridgeProxy":              expected.L1StandardBridgeProxy.Hex(),
		"L1CrossDomainMessengerProxy":        expected.L1CrossDomainMessengerProxy.Hex(),
		"EthLockboxProxy":                    expected.EthLockboxProxy.Hex(),
		"OptimismPortalProxy":                expected.OptimismPortalProxy.Hex(),
		"DisputeGameFactoryProxy":            expected.DisputeGameFactoryProxy.Hex(),
		"AnchorStateRegistryProxy":           expected.AnchorStateRegistryProxy.Hex(),
		"FaultDisputeGameImpl":               expected.FaultDisputeGame.Hex(),
		"PermissionedDisputeGameImpl":        expected.PermissionedDisputeGame.Hex(),
		"DelayedWethPermissionedGameProxy":   expected.DelayedWETHPermissionedGameProxy.Hex(),
		"DelayedWethPermissionlessGameProxy": expected.DelayedWETHPermissionlessGameProxy.Hex(),
	}
	state := deployer.OpaqueState{"opChainDeployments": []any{deployment}}

	actual, err := DeployOutputFromState(state, 0)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestValidateDeployOutputInReceipt(t *testing.T) {
	out := testV7DeployOutput()
	logs := []*types.Log{
		{Address: out.OpChainProxyAdmin},
		{Address: out.AddressManager},
		{Address: out.L1ERC721BridgeProxy},
		{Address: out.SystemConfigProxy},
		{Address: out.OptimismMintableERC20FactoryProxy},
		{Address: out.L1StandardBridgeProxy},
		{Address: out.L1CrossDomainMessengerProxy},
		{Address: out.EthLockboxProxy},
		{Address: out.OptimismPortalProxy},
		{Address: out.DisputeGameFactoryProxy},
		{Address: out.AnchorStateRegistryProxy},
		{Address: out.DelayedWETHPermissionedGameProxy},
	}

	require.NoError(t, ValidateDeployOutputInReceipt(logs, out))

	logs = logs[:3]
	require.ErrorContains(t, ValidateDeployOutputInReceipt(logs, out), "SystemConfigProxy")
}
