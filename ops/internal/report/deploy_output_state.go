package report

import (
	"fmt"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// DeployOutputFromState reads the per-chain addresses needed by the L1 scanner.
// OPCM V2 no longer emits the aggregate Deployed event used by older releases.
func DeployOutputFromState(st deployer.OpaqueState, idx int) (DeployOPChainOutput, error) {
	var out DeployOPChainOutput
	reads := []struct {
		name string
		dst  *common.Address
		read func(int) (common.Address, error)
	}{
		{"OpChainProxyAdmin", &out.OpChainProxyAdmin, st.ReadProxyAdminImpl},
		{"AddressManager", &out.AddressManager, st.ReadAddressManagerImpl},
		{"L1ERC721BridgeProxy", &out.L1ERC721BridgeProxy, st.ReadL1Erc721BridgeProxy},
		{"SystemConfigProxy", &out.SystemConfigProxy, st.ReadSystemConfigProxy},
		{"OptimismMintableERC20FactoryProxy", &out.OptimismMintableERC20FactoryProxy, st.ReadOptimismMintableErc20FactoryProxy},
		{"L1StandardBridgeProxy", &out.L1StandardBridgeProxy, st.ReadL1StandardBridgeProxy},
		{"L1CrossDomainMessengerProxy", &out.L1CrossDomainMessengerProxy, st.ReadL1CrossDomainMessengerProxy},
		{"EthLockboxProxy", &out.EthLockboxProxy, st.ReadEthLockboxProxy},
		{"OptimismPortalProxy", &out.OptimismPortalProxy, st.ReadOptimismPortalProxy},
		{"DisputeGameFactoryProxy", &out.DisputeGameFactoryProxy, st.ReadDisputeGameFactoryProxy},
		{"AnchorStateRegistryProxy", &out.AnchorStateRegistryProxy, st.ReadAnchorStateRegistryProxy},
		{"FaultDisputeGame", &out.FaultDisputeGame, st.ReadFaultDisputeGameImpl},
		{"PermissionedDisputeGame", &out.PermissionedDisputeGame, st.ReadPermissionedDisputeGameImpl},
		{"DelayedWETHPermissionedGameProxy", &out.DelayedWETHPermissionedGameProxy, st.ReadDelayedWethPermissionedGameProxy},
		{"DelayedWETHPermissionlessGameProxy", &out.DelayedWETHPermissionlessGameProxy, st.ReadDelayedWethPermissionlessGameProxy},
	}

	for _, item := range reads {
		addr, err := item.read(idx)
		if err != nil {
			return out, fmt.Errorf("failed to read %s from state: %w", item.name, err)
		}
		*item.dst = addr
	}

	return out, nil
}

// ValidateDeployOutputInReceipt ties state-derived addresses to the deployment
// transaction. Each per-chain contract must have emitted at least one log in
// the successful receipt before the scanner trusts the state fallback.
func ValidateDeployOutputInReceipt(logs []*types.Log, out DeployOPChainOutput) error {
	emitters := make(map[common.Address]struct{}, len(logs))
	for _, log := range logs {
		emitters[log.Address] = struct{}{}
	}

	addresses := []struct {
		name string
		addr common.Address
	}{
		{"OpChainProxyAdmin", out.OpChainProxyAdmin},
		{"AddressManager", out.AddressManager},
		{"L1ERC721BridgeProxy", out.L1ERC721BridgeProxy},
		{"SystemConfigProxy", out.SystemConfigProxy},
		{"OptimismMintableERC20FactoryProxy", out.OptimismMintableERC20FactoryProxy},
		{"L1StandardBridgeProxy", out.L1StandardBridgeProxy},
		{"L1CrossDomainMessengerProxy", out.L1CrossDomainMessengerProxy},
		{"EthLockboxProxy", out.EthLockboxProxy},
		{"OptimismPortalProxy", out.OptimismPortalProxy},
		{"DisputeGameFactoryProxy", out.DisputeGameFactoryProxy},
		{"AnchorStateRegistryProxy", out.AnchorStateRegistryProxy},
		{"DelayedWETHPermissionedGameProxy", out.DelayedWETHPermissionedGameProxy},
	}

	for _, item := range addresses {
		if item.addr == (common.Address{}) {
			return fmt.Errorf("%s is the zero address", item.name)
		}
		if _, ok := emitters[item.addr]; !ok {
			return fmt.Errorf("%s address %s did not emit a deployment log", item.name, item.addr)
		}
	}

	return nil
}
