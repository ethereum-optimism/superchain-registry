package state

import (
	_ "embed"
	"fmt"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer/opaque_map"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/tomwright/dasel"
)

//go:embed configs/v1-state.json
var standardV1State []byte

//go:embed configs/v1-intent.toml
var standardV1Intent []byte

func MergeStateV1(userState OpaqueState) (opaque_map.OpaqueMap, OpaqueState, error) {
	l1ChainID, err := userState.ReadL1ChainID()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read L1 chain ID: %w", err)
	}
	stdIntent, err := StandardIntentV1(l1ChainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create standard intent: %w", err)
	}
	stdState, err := StandardStateV1(l1ChainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create standard state: %w", err)
	}
	return mergeStateV2(userState, stdIntent, stdState)
}

func StandardStateV1(l1ChainID uint64) (OpaqueState, error) {
	return standardState(l1ChainID, validation.Semver180, standardV1State)
}

func StandardIntentV1(l1ChainID uint64) (opaque_map.OpaqueMap, error) {
	return standardIntentV1(l1ChainID, standardV1Intent)
}

func standardIntentV1(l1ChainID uint64, data []byte) (opaque_map.OpaqueMap, error) {
	intent, err := standardIntentV2(l1ChainID, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create standard intent: %w", err)
	}

	root := dasel.New(intent)
	// This is a hack to workaround an op-deployer bug where the protocolVersionsOwner is incorrectly
	// set to the protocolVersionsImpl address. So we mirror that value here so we can pass the intent validation.
	mustPutString(root, "superchainRoles.protocolVersionsOwner", stringWrapper("0x79ADD5713B383DAa0a138d3C4780C7A1804a8090"))

	return intent, nil
}
