package deployer

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tomwright/dasel"
)

type SchemaVersion string

var (
	unknownVersion SchemaVersion = "unknown"
	V3             SchemaVersion = "v3.0.0"
	V2             SchemaVersion = "v2.0.0"
)

func GetSchemaVersion(st OpaqueMapping) SchemaVersion {
	node, err := dasel.New(st).Query("appliedIntent.l1ContractsLocator")
	if err != nil {
		return unknownVersion
	}

	l1ContractsLocator, ok := node.InterfaceValue().(string)
	if !ok {
		return unknownVersion
	}

	switch l1ContractsLocator {
	case "tag://op-contracts/v3.0.0-rc.2":
		return V3
	case "tag://op-contracts/v2.0.0":
		return V2
	default:
		return unknownVersion
	}
}

// GetNumChains returns the number of chains in the supplied state file, or
// an error if the state file is not of a supported schema.
func GetNumChains(st OpaqueMapping) (int, error) {
	switch GetSchemaVersion(st) {
	case V3, V2:
		return getNumChainsV2(st)
	default:
		return 0, fmt.Errorf("unsupported schema version")
	}
}

func getNumChainsV2(st OpaqueMapping) (int, error) {
	// So far the below is expected to work for all known schema versions
	node, err := dasel.New(st).Query("appliedIntent.chains")
	if err != nil {
		return 0, fmt.Errorf("failed to query appliedIntent.chains: %w", err)
	}
	slice, ok := node.InterfaceValue().([]interface{})
	if !ok {
		return 0, errors.New("failed to cast appliedIntent.chains to []interface{}")
	}
	return len(slice), nil
}

func GetChainID(st OpaqueMapping, idx int) (common.Hash, error) {
	switch GetSchemaVersion(st) {
	case V3, V2:
		return getChainIDV2(st, idx)
	default:
		return common.Hash{}, fmt.Errorf("unsupported schema version")
	}
}
func getChainIDV2(st OpaqueMapping, idx int) (common.Hash, error) {
	node, err := dasel.New(st).Query(fmt.Sprintf("appliedIntent.chains.[%d].id", idx))
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to query chain ID %d: %w", idx, err)
	}
	chainIDStr, ok := node.InterfaceValue().(string)
	if !ok {
		return common.Hash{}, fmt.Errorf("failed to cast chain ID %d to string", idx)
	}
	return common.HexToHash(chainIDStr), nil

}
