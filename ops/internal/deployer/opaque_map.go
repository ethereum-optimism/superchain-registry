package deployer

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tomwright/dasel"
)

type (
	OpaqueMap   map[string]any
	OpaqueState OpaqueMap
)

// useInts converts all float64 values without fractional parts to int64 values in a map
// so that they are properly marshaled to TOML
func useInts(m map[string]any) {
	for k, v := range m {
		switch val := v.(type) {
		case float64:
			// If the float has no fractional part, convert to int
			if val == float64(int64(val)) {
				m[k] = int64(val)
			}
		case map[string]any:
			// Recursively process nested maps
			useInts(val)
		case []any:
			// Process arrays
			for i, item := range val {
				if fItem, ok := item.(float64); ok && fItem == float64(int64(fItem)) {
					val[i] = int64(fItem)
				} else if mapItem, ok := item.(map[string]any); ok {
					useInts(mapItem)
				}
			}
		}
	}
}

func (om OpaqueState) ReadL1ChainID() (uint64, error) {
	node := dasel.New(om)
	l1ChainIDNode, err := node.Query("appliedIntent.l1ChainID")
	if err != nil {
		return 0, fmt.Errorf("failed to read L1 chain ID: %w", err)
	}
	l1ChainIDFloat, ok := l1ChainIDNode.InterfaceValue().(float64)
	if !ok {
		return 0, errors.New("failed to parse L1 chain ID")
	}
	return uint64(l1ChainIDFloat), nil
}

func (om OpaqueState) ReadL1ContractsLocator() (string, error) {
	node := dasel.New(om)
	l1ContractsReleaseNode, err := node.Query("appliedIntent.l1ContractsLocator")
	if err != nil {
		return "", fmt.Errorf("failed to read L1 contracts release: %w", err)
	}
	l1ContractsRelease, ok := l1ContractsReleaseNode.InterfaceValue().(string)
	if !ok {
		return "", errors.New("failed to parse L1 contracts release")
	}
	return l1ContractsRelease, nil
}

func (om OpaqueState) ReadL2ContractsLocator() (string, error) {
	node := dasel.New(om)
	l2ContractsReleaseNode, err := node.Query("appliedIntent.l2ContractsLocator")
	if err != nil {
		return "", fmt.Errorf("failed to read L2 contracts release: %w", err)
	}
	l2ContractsRelease, ok := l2ContractsReleaseNode.InterfaceValue().(string)
	if !ok {
		return "", errors.New("failed to parse L2 contracts release")
	}
	return l2ContractsRelease, nil
}

func (om OpaqueState) ReadL2ChainId(idx int) (string, error) {
	node := dasel.New(om)
	l2ChainIdNode, err := node.Query(fmt.Sprintf("appliedIntent.chains.[%d].id", idx))
	if err != nil {
		return "", fmt.Errorf("failed to read L2 chain ID: %w", err)
	}
	l2ChainId, ok := l2ChainIdNode.InterfaceValue().(string)
	if !ok {
		return "", errors.New("failed to parse L2 chain ID")
	}
	return l2ChainId, nil
}

func (om OpaqueState) ReadSystemConfigProxy(idx int) (common.Address, error) {
	node := dasel.New(om)
	systemConfigProxyNode, err := node.Query(fmt.Sprintf("appliedIntent.opChainDeployments.[%d].SystemConfigProxy", idx))
	if err == nil {
		systemConfigProxy, ok := systemConfigProxyNode.InterfaceValue().(string)
		if !ok {
			return common.Address{}, errors.New("failed to parse SystemConfigProxy")
		}
		return common.HexToAddress(systemConfigProxy), nil
	}

	// Fallback to the legacy field name
	systemConfigProxyNode, err = node.Query(fmt.Sprintf("appliedIntent.opChainDeployments.[%d].systemConfigProxyAddress", idx))
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to read SystemConfigProxy: %w", err)
	}

	systemConfigProxy, ok := systemConfigProxyNode.InterfaceValue().(string)
	if !ok {
		return common.Address{}, errors.New("failed to parse SystemConfigProxy")
	}
	return common.HexToAddress(systemConfigProxy), nil
}

func (om OpaqueState) ReadL1StandardBridgeProxy(idx int) (common.Address, error) {
	node := dasel.New(om)
	l1StandardBridgeProxyNode, err := node.Query(fmt.Sprintf("appliedIntent.opChainDeployments.[%d].L1StandardBridgeProxy", idx))
	if err == nil {
		l1StandardBridgeProxy, ok := l1StandardBridgeProxyNode.InterfaceValue().(string)
		if !ok {
			return common.Address{}, errors.New("failed to parse L1StandardBridgeProxy")
		}
		return common.HexToAddress(l1StandardBridgeProxy), nil
	}

	// Fallback to the legacy field name
	l1StandardBridgeProxyNode, err = node.Query(fmt.Sprintf("appliedIntent.opChainDeployments.[%d].l1StandardBridgeProxyAddress", idx))
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to read L1StandardBridgeProxy: %w", err)
	}

	l1StandardBridgeProxy, ok := l1StandardBridgeProxyNode.InterfaceValue().(string)
	if !ok {
		return common.Address{}, errors.New("failed to parse L1StandardBridgeProxy")
	}
	return common.HexToAddress(l1StandardBridgeProxy), nil
}
