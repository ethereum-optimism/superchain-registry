package state

import (
	"fmt"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer/opaque_map"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tomwright/dasel"
)

type OpaqueState opaque_map.OpaqueMap

// QueryOpaqueState queries the OpaqueState for the given paths in order,
// and returns the first successful result (and an error otherwise)
func QueryOpaqueState[T any](om OpaqueState, paths ...string) (T, error) {
	node := dasel.New(om)
	resultNode := new(dasel.Node)
	var err error
	for _, path := range paths {
		resultNode, err = node.Query(path)
		if err == nil {
			break
		}
	}

	var zero T
	if err != nil {
		return zero, fmt.Errorf("failed to successfully query any of the paths: %w", err)
	}

	result, ok := resultNode.InterfaceValue().(T)
	if !ok {
		return zero, fmt.Errorf("failed to assert type of result %T as %T", resultNode.InterfaceValue(), zero)
	}
	return result, nil
}

// queryString retrieves a string value from the given path
func (om OpaqueState) queryString(paths ...string) (string, error) {
	return QueryOpaqueState[string](om, paths...)
}

// queryAddress retrieves an address from the given path, with an optional fallback path
func (om OpaqueState) queryAddress(paths ...string) (common.Address, error) {
	val, err := QueryOpaqueState[string](om, paths...)
	if err != nil {
		return common.Address{}, err
	}
	return common.HexToAddress(val), nil
}

// queryInt retrieves a uint64 value from the given path
func (om OpaqueState) queryInt(path string) (uint64, error) {
	node := dasel.New(om)
	resultNode, err := node.Query(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read path %s: %w", path, err)
	}

	// Handle both float64 and int64 cases since the JSON parser might give us either
	switch val := resultNode.InterfaceValue().(type) {
	case float64:
		if val != float64(uint64(val)) {
			return 0, fmt.Errorf("value at path %s has fractional part", path)
		}
		return uint64(val), nil
	case int64:
		if val < 0 {
			return 0, fmt.Errorf("value at path %s is negative", path)
		}
		return uint64(val), nil
	default:
		return 0, fmt.Errorf("failed to parse integer at path %s", path)
	}
}

func (om OpaqueState) ReadL1ChainID() (uint64, error) {
	return om.queryInt("appliedIntent.l1ChainID")
}

func (om OpaqueState) ReadL1ContractsLocator() (string, error) {
	return om.queryString("appliedIntent.l1ContractsLocator")
}

func (om OpaqueState) ReadL2ContractsLocator() (string, error) {
	return om.queryString("appliedIntent.l2ContractsLocator")
}

func (om OpaqueState) ReadL2ChainId(idx int) (string, error) {
	return om.queryString(fmt.Sprintf("appliedIntent.chains.[%d].id", idx))
}

func (om OpaqueState) ReadSystemConfigProxy(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("opChainDeployments.[%d].SystemConfigProxy", idx),
		fmt.Sprintf("opChainDeployments.[%d].systemConfigProxyAddress", idx),
	)
}

func (om OpaqueState) ReadL1StandardBridgeProxy(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("opChainDeployments.[%d].L1StandardBridgeProxy", idx),
		fmt.Sprintf("opChainDeployments.[%d].l1StandardBridgeProxyAddress", idx),
	)
}

func (om OpaqueState) ReadAddressManagerImpl(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("opChainDeployments.[%d].AddressManagerImpl", idx),
		fmt.Sprintf("opChainDeployments.[%d].addressManagerAddress", idx),
	)
}

func (om OpaqueState) ReadOptimismPortalProxy(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("opChainDeployments.[%d].OptimismPortalProxy", idx),
		fmt.Sprintf("opChainDeployments.[%d].optimismPortalProxyAddress", idx),
	)
}

func (om OpaqueState) ReadL1CrossDomainMessengerProxy(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("opChainDeployments.[%d].L1CrossDomainMessengerProxy", idx),
		fmt.Sprintf("opChainDeployments.[%d].l1CrossDomainMessengerProxyAddress", idx),
	)
}

func (om OpaqueState) ReadOptimismMintableErc20FactoryProxy(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("opChainDeployments.[%d].OptimismMintableErc20FactoryProxy", idx),
		fmt.Sprintf("opChainDeployments.[%d].optimismMintableERC20FactoryProxyAddress", idx),
	)
}

func (om OpaqueState) ReadProxyAdminImpl(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("opChainDeployments.[%d].OpChainProxyAdminImpl", idx),
		fmt.Sprintf("opChainDeployments.[%d].proxyAdminAddress", idx),
	)
}

func (om OpaqueState) ReadL1Erc721BridgeProxy(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("opChainDeployments.[%d].L1Erc721BridgeProxy", idx),
		fmt.Sprintf("opChainDeployments.[%d].l1ERC721BridgeProxyAddress", idx),
	)
}

func (om OpaqueState) ReadAnchorStateRegistryProxy(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("opChainDeployments.[%d].AnchorStateRegistryProxy", idx),
		fmt.Sprintf("opChainDeployments.[%d].anchorStateRegistryProxyAddress", idx),
	)
}

func (om OpaqueState) ReadDelayedWethPermissionedGameProxy(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("opChainDeployments.[%d].DelayedWethPermissionedGameProxy", idx),
		fmt.Sprintf("opChainDeployments.[%d].delayedWETHPermissionedGameProxyAddress", idx),
	)
}

func (om OpaqueState) ReadDisputeGameFactoryProxy(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("opChainDeployments.[%d].DisputeGameFactoryProxy", idx),
		fmt.Sprintf("opChainDeployments.[%d].disputeGameFactoryProxyAddress", idx),
	)
}

func (om OpaqueState) ReadPermissionedDisputeGameImpl(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("opChainDeployments.[%d].PermissionedDisputeGameImpl", idx),
		fmt.Sprintf("opChainDeployments.[%d].permissionedDisputeGameAddress", idx),
	)
}

func (om OpaqueState) ReadProtocolVersionsProxy() (common.Address, error) {
	return om.queryAddress(
		"superchainContracts.ProtocolVersionsProxy",
		"superchainDeployment.protocolVersionsProxyAddress",
	)
}

func (om OpaqueState) ReadSuperchainConfigProxy() (common.Address, error) {
	return om.queryAddress(
		"superchainContracts.SuperchainConfigProxy",
		"superchainDeployment.superchainConfigProxyAddress",
	)
}

func (om OpaqueState) ReadOpcmImpl() (common.Address, error) {
	return om.queryAddress(
		"implementationsDeployment.OpcmImpl",
		"implementationsDeployment.opcmAddress",
	)
}

func (om OpaqueState) ReadSystemConfigOwner(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("appliedIntent.chains.[%d].roles.systemConfigOwner", idx),
	)
}

func (om OpaqueState) ReadProxyAdminOwner(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("appliedIntent.chains.[%d].roles.l1ProxyAdminOwner", idx),
	)
}

func (om OpaqueState) ReadGuardian(idx int) (common.Address, error) {
	return om.queryAddress(
		"appliedIntent.superchainRoles.guardian",
		"superchainRoles.SuperchainGuardian",
	)
}

func (om OpaqueState) ReadChallenger(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("appliedIntent.chains.[%d].roles.challenger", idx),
	)
}

func (om OpaqueState) ReadProposer(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("appliedIntent.chains.[%d].roles.proposer", idx),
	)
}

func (om OpaqueState) ReadUnsafeBlockSigner(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("appliedIntent.chains.[%d].roles.unsafeBlockSigner", idx),
	)
}

func (om OpaqueState) ReadBatchSubmitter(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("appliedIntent.chains.[%d].roles.batcher", idx),
	)
}

func (om OpaqueState) GetNumChains() (int, error) {
	return QueryOpaqueState[int](om, "appliedIntent.chains.[#]")
}

func (om OpaqueState) GetChainID(idx int) (uint64, error) {
	val, err := om.queryString(fmt.Sprintf("appliedIntent.chains.[%d].id", idx))
	if err != nil {
		return 0, fmt.Errorf("failed to read chain id: %w", err)
	}
	h := common.HexToHash(val)
	return h.Big().Uint64(), nil
}

func (s OpaqueState) ReadInteropDepSet() (map[string]interface{}, error) {
	interopDepSet, ok := s["interopDepSet"]
	if !ok || interopDepSet == nil {
		return nil, nil
	}

	depSet, ok := interopDepSet.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("interopDepSet is not a map[string]interface{}")
	}

	return depSet, nil
}
