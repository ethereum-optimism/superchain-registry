package deployer

import (
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

// queryString retrieves a string value from the given path
func (om OpaqueState) queryString(path string) (string, error) {
	node := dasel.New(om)
	resultNode, err := node.Query(path)
	if err != nil {
		return "", fmt.Errorf("failed to read path %s: %w", path, err)
	}
	result, ok := resultNode.InterfaceValue().(string)
	if !ok {
		return "", fmt.Errorf("failed to parse string at path %s", path)
	}
	return result, nil
}

// queryAddress retrieves an address from the given path, with an optional fallback path
func (om OpaqueState) queryAddress(path, fallbackPath string) (common.Address, error) {
	node := dasel.New(om)
	resultNode, err := node.Query(path)
	if err == nil {
		if result, ok := resultNode.InterfaceValue().(string); ok {
			return common.HexToAddress(result), nil
		}
		return common.Address{}, fmt.Errorf("failed to parse address at path %s", path)
	}

	if fallbackPath != "" {
		resultNode, err = node.Query(fallbackPath)
		if err != nil {
			return common.Address{}, fmt.Errorf("failed to read address at both paths %s and %s: %w", path, fallbackPath, err)
		}
		if result, ok := resultNode.InterfaceValue().(string); ok {
			return common.HexToAddress(result), nil
		}
	}

	return common.Address{}, fmt.Errorf("failed to parse address at path %s", fallbackPath)
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
		fmt.Sprintf("appliedIntent.opChainDeployments.[%d].SystemConfigProxy", idx),
		fmt.Sprintf("appliedIntent.opChainDeployments.[%d].systemConfigProxyAddress", idx),
	)
}

func (om OpaqueState) ReadL1StandardBridgeProxy(idx int) (common.Address, error) {
	return om.queryAddress(
		fmt.Sprintf("appliedIntent.opChainDeployments.[%d].L1StandardBridgeProxy", idx),
		fmt.Sprintf("appliedIntent.opChainDeployments.[%d].l1StandardBridgeProxyAddress", idx),
	)
}
