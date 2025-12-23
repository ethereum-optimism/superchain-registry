package state

import (
	_ "embed"
	"fmt"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer/opaque_map"
	"github.com/ethereum-optimism/superchain-registry/validation"
)

//go:embed configs/v3-state.json
var standardV3State []byte

//go:embed configs/v3-intent.toml
var standardV3Intent []byte

func MergeStateV3(userState OpaqueState) (opaque_map.OpaqueMap, OpaqueState, error) {
	l1ChainID, err := userState.ReadL1ChainID()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read L1 chain ID: %w", err)
	}
	stdIntent, err := StandardIntentV3(l1ChainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create standard intent: %w", err)
	}
	stdState, err := StandardStateV3(l1ChainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create standard state: %w", err)
	}
	// V2 is correct here. V3's state is the same as V2, except with a
	// slightly different intent that contains operator fee fields.
	return mergeStateV2(userState, stdIntent, stdState)
}

func StandardStateV3(l1ChainID uint64) (OpaqueState, error) {
	return standardState(l1ChainID, validation.Semver300, standardV3State)
}

func StandardIntentV3(l1ChainID uint64) (opaque_map.OpaqueMap, error) {
	return standardIntentV2(l1ChainID, standardV3Intent)
}
