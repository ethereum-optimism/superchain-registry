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

// QueryOpaqueMap queries the OpaqueState for the given paths in order,
// and returns the first successful result (and an error otherwise)
func QueryOpaqueMap[T any](om OpaqueState, paths ...string) (T, error) {
	node := dasel.New(om)
	resultNode := new(dasel.Node)
	err := fmt.Errorf("not found")
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
	return QueryOpaqueMap[string](om, paths...)
}

// queryAddress retrieves an address from the given path, with an optional fallback path
func (om OpaqueState) queryAddress(paths ...string) (common.Address, error) {
	return QueryOpaqueMap[common.Address](om, paths...)
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

func (om OpaqueState) ReadProtocolVersionsProxy() (common.Address, error) {
	return om.queryAddress(
		"superchainDeployments.ProtocolVersionsProxyAddress",
		"superchainContracts.ProtocolVersionsProxy",
	)
}

func (om OpaqueState) ReadSuperchainConfigProxy() (common.Address, error) {
	return om.queryAddress(
		"superchainDeployments.SuperchainConfigProxyAddress",
		"superchainContracts.SuperchainConfigProxy",
	)
}

func (om OpaqueState) ReadOpcmAddress() (common.Address, error) {
	return om.queryAddress(
		"implementationsDeployment.OpcmAddress",
		"implementationsDeployment.OpcmImpl",
	)
}

func (om OpaqueState) GetNumChains() (int, error) {
	val, err := dasel.New(om).Query("appliedIntent.chains")
	if err != nil {
		return 0, fmt.Errorf("failed to read number of chains: %w", err)
	}
	chains, ok := val.InterfaceValue().([]any)
	if !ok {
		return 0, fmt.Errorf("failed to parse chains")
	}
	return len(chains), nil
}

func (om OpaqueState) GetChainID(idx int) (uint64, error) {
	val, err := om.queryString(fmt.Sprintf("appliedIntent.chains.[%d].id", idx))
	if err != nil {
		return 0, fmt.Errorf("failed to read chain id: %w", err)
	}
	h := common.HexToHash(val)
	return h.Big().Uint64(), nil
}
