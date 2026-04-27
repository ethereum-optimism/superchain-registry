package report

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3"
)

type DeployedEvent struct {
	OutputVersion *big.Int
	L2ChainID     common.Hash
	Deployer      common.Address
	DeployOutput  DeployOPChainOutput
}

type DeployOPChainOutput struct {
	OpChainProxyAdmin                  common.Address
	AddressManager                     common.Address
	L1ERC721BridgeProxy                common.Address
	SystemConfigProxy                  common.Address
	OptimismMintableERC20FactoryProxy  common.Address
	L1StandardBridgeProxy              common.Address
	L1CrossDomainMessengerProxy        common.Address
	EthLockboxProxy                    common.Address
	OptimismPortalProxy                common.Address
	DisputeGameFactoryProxy            common.Address
	AnchorStateRegistryProxy           common.Address
	AnchorStateRegistryImpl            common.Address
	FaultDisputeGame                   common.Address
	PermissionedDisputeGame            common.Address
	DelayedWETHPermissionedGameProxy   common.Address
	DelayedWETHPermissionlessGameProxy common.Address
}

var (
	deployedEventABI_v1 = w3.MustNewEvent(`Deployed(uint256 indexed, uint256 indexed, address indexed, bytes)`)
	deployedEventABI_v2 = w3.MustNewEvent(`Deployed(uint256 indexed, address indexed, bytes)`)

	deployOutputEvABI_v1 = w3.MustNewFunc(`
dummy(
	address opChainProxyAdmin,
	address addressManager,
	address l1ERC721BridgeProxy,
	address systemConfigProxy,
	address optimismMintableERC20FactoryProxy,
	address l1StandardBridgeProxy,
	address l1CrossDomainMessengerProxy,
	address optimismPortalProxy,
	address disputeGameFactoryProxy,
	address anchorStateRegistryProxy,
	address anchorStateRegistryImpl,
	address faultDisputeGame,
	address permissionedDisputeGame,
	address delayedWETHPermissionedGameProxy,
	address delayedWETHPermissionlessGameProxy
)
`, "")

	deployOutputEvABI_v2 = w3.MustNewFunc(`
dummy(
	address opChainProxyAdmin,
	address addressManager,
	address l1ERC721BridgeProxy,
	address systemConfigProxy,
	address optimismMintableERC20FactoryProxy,
	address l1StandardBridgeProxy,
	address l1CrossDomainMessengerProxy,
	address ethLockboxProxy,
	address optimismPortalProxy,
	address disputeGameFactoryProxy,
	address anchorStateRegistryProxy,
	address faultDisputeGame,
	address permissionedDisputeGame,
	address delayedWETHPermissionedGameProxy,
	address delayedWETHPermissionlessGameProxy
)
`, "")
)

// ParseDeployedEvent scans receipt logs for a Deployed event and decodes it.
// Supports both old (<= op-contracts/v4.0) and new (>= op-contracts/v4.1+) event signatures.
func ParseDeployedEvent(logs []*types.Log) (*DeployedEvent, error) {
	var deploymentLog *types.Log
	var useNewSignature bool

	// Scan for Deployed event, looking for any matching event signature
	for _, log := range logs {
		// Try new signature first (v4.1+)
		if log.Topics[0] == deployedEventABI_v2.Topic0 {
			if deploymentLog != nil {
				return nil, fmt.Errorf("multiple Deployed events in receipt, this is unsupported")
			}
			deploymentLog = log
			useNewSignature = true
			continue
		}

		// Fall back to old signature (v4.0 and earlier)
		if log.Topics[0] == deployedEventABI_v1.Topic0 {
			if deploymentLog != nil {
				return nil, fmt.Errorf("multiple Deployed events in receipt, this is unsupported")
			}
			deploymentLog = log
			useNewSignature = false
			continue
		}
	}

	if deploymentLog == nil {
		return nil, fmt.Errorf("no Deployed event in receipt")
	}

	if useNewSignature {
		return decodeDeployedEventV2(deploymentLog)
	} else {
		return decodeDeployedEventV1(deploymentLog)
	}
}

// decodeDeployedEventV1 decodes OPCM Deployed event (op-contracts/v4.0 and earlier)
func decodeDeployedEventV1(log *types.Log) (*DeployedEvent, error) {
	outVersion := new(big.Int)
	var l2ChainID *big.Int
	var deployer common.Address
	var deployOutput []byte

	if err := deployedEventABI_v1.DecodeArgs(log, &outVersion, &l2ChainID, &deployer, &deployOutput); err != nil {
		return nil, fmt.Errorf("failed to decode v1 Deployed event: %w", err)
	}

	deployedEv := &DeployedEvent{
		OutputVersion: outVersion,
		L2ChainID:     common.BigToHash(l2ChainID),
		Deployer:      deployer,
	}

	deployOutputWithSel := make([]byte, 4+len(deployOutput))
	copy(deployOutputWithSel, deployOutputEvABI_v1.Selector[:])
	copy(deployOutputWithSel[4:], deployOutput)

	if err := deployOutputEvABI_v1.DecodeArgs(
		deployOutputWithSel,
		&deployedEv.DeployOutput.OpChainProxyAdmin,
		&deployedEv.DeployOutput.AddressManager,
		&deployedEv.DeployOutput.L1ERC721BridgeProxy,
		&deployedEv.DeployOutput.SystemConfigProxy,
		&deployedEv.DeployOutput.OptimismMintableERC20FactoryProxy,
		&deployedEv.DeployOutput.L1StandardBridgeProxy,
		&deployedEv.DeployOutput.L1CrossDomainMessengerProxy,
		&deployedEv.DeployOutput.OptimismPortalProxy,
		&deployedEv.DeployOutput.DisputeGameFactoryProxy,
		&deployedEv.DeployOutput.AnchorStateRegistryProxy,
		&deployedEv.DeployOutput.AnchorStateRegistryImpl,
		&deployedEv.DeployOutput.FaultDisputeGame,
		&deployedEv.DeployOutput.PermissionedDisputeGame,
		&deployedEv.DeployOutput.DelayedWETHPermissionedGameProxy,
		&deployedEv.DeployOutput.DelayedWETHPermissionlessGameProxy,
	); err != nil {
		return nil, fmt.Errorf("failed to decode deploy output: %w", err)
	}
	return deployedEv, nil
}

// decodeDeployedEventV2 decodes OPCM Deployed event (op-contracts/v4.1+)
func decodeDeployedEventV2(log *types.Log) (*DeployedEvent, error) {
	var l2ChainID *big.Int
	var deployer common.Address
	var deployOutput []byte

	if err := deployedEventABI_v2.DecodeArgs(log, &l2ChainID, &deployer, &deployOutput); err != nil {
		return nil, fmt.Errorf("failed to decode v2 Deployed event: %w", err)
	}

	deployedEv := &DeployedEvent{
		OutputVersion: nil, // v2 event doesn't include OutputVersion
		L2ChainID:     common.BigToHash(l2ChainID),
		Deployer:      deployer,
	}

	deployOutputWithSel := make([]byte, 4+len(deployOutput))
	copy(deployOutputWithSel, deployOutputEvABI_v2.Selector[:])
	copy(deployOutputWithSel[4:], deployOutput)

	if err := deployOutputEvABI_v2.DecodeArgs(
		deployOutputWithSel,
		&deployedEv.DeployOutput.OpChainProxyAdmin,
		&deployedEv.DeployOutput.AddressManager,
		&deployedEv.DeployOutput.L1ERC721BridgeProxy,
		&deployedEv.DeployOutput.SystemConfigProxy,
		&deployedEv.DeployOutput.OptimismMintableERC20FactoryProxy,
		&deployedEv.DeployOutput.L1StandardBridgeProxy,
		&deployedEv.DeployOutput.L1CrossDomainMessengerProxy,
		&deployedEv.DeployOutput.EthLockboxProxy,
		&deployedEv.DeployOutput.OptimismPortalProxy,
		&deployedEv.DeployOutput.DisputeGameFactoryProxy,
		&deployedEv.DeployOutput.AnchorStateRegistryProxy,
		&deployedEv.DeployOutput.FaultDisputeGame,
		&deployedEv.DeployOutput.PermissionedDisputeGame,
		&deployedEv.DeployOutput.DelayedWETHPermissionedGameProxy,
		&deployedEv.DeployOutput.DelayedWETHPermissionlessGameProxy,
	); err != nil {
		return nil, fmt.Errorf("failed to decode deploy output: %w", err)
	}

	return deployedEv, nil
}
