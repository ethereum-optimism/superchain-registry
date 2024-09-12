package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLegacyPlasmaConfig_error(t *testing.T) {
	jsonData := `{
		"block_time": 2,
		"da_commitment_type": "GenericCommitment"
	}`

	var plasmaConfig LegacyPlasma
	err := json.Unmarshal([]byte(jsonData), &plasmaConfig)
	require.NoError(t, err, "Error unmarshaling JSON: %v", err)

	require.Error(t, plasmaConfig.CheckNonNilFields(), "Expected an error due to non-nil fields, but got nil")
}

func TestLegacyPlasmaConfig_success(t *testing.T) {
	jsonData := `{
		"block_time": 2
	}`

	var plasmaConfig LegacyPlasma
	err := json.Unmarshal([]byte(jsonData), &plasmaConfig)
	require.NoError(t, err, "Error unmarshaling JSON: %v", err)

	require.NoError(t, plasmaConfig.CheckNonNilFields(), "Expected no error when checking for non nil plasma fields")
}
