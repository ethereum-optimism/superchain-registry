package manage

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
)

func TestCopyDeployConfigHFTimes(t *testing.T) {
	a := &genesis.UpgradeScheduleDeployConfig{
		L2GenesisCanyonTimeOffset: new(hexutil.Uint64),
		L2GenesisDeltaTimeOffset:  new(hexutil.Uint64),
	}
	*a.L2GenesisCanyonTimeOffset = hexutil.Uint64(1)
	*a.L2GenesisDeltaTimeOffset = hexutil.Uint64(2)

	b := &config.Hardforks{}

	require.NoError(t, CopyDeployConfigHFTimes(a, b))
	require.Equal(t, &config.Hardforks{
		CanyonTime: config.NewHardforkTime(1),
		DeltaTime:  config.NewHardforkTime(2),
	}, b)
}

func TestExtractInteropDepSet(t *testing.T) {
	tests := []struct {
		name        string
		stateData   deployer.OpaqueState
		expected    *config.Interop
		expectError bool
	}{
		{
			name: "valid interop dependencies",
			stateData: deployer.OpaqueState{
				"interopDepSet": map[string]interface{}{
					"dependencies": map[string]interface{}{
						"123": map[string]interface{}{},
						"456": map[string]interface{}{},
					},
				},
			},
			expected: &config.Interop{
				Dependencies: map[string]config.StaticConfigDependency{
					"123": {},
					"456": {},
				},
			},
			expectError: false,
		},
		{
			name:        "nil interopDepSet",
			stateData:   deployer.OpaqueState{},
			expected:    nil,
			expectError: false,
		},
		{
			name: "empty dependencies",
			stateData: deployer.OpaqueState{
				"interopDepSet": map[string]interface{}{
					"dependencies": map[string]interface{}{},
				},
			},
			expected: &config.Interop{
				Dependencies: map[string]config.StaticConfigDependency{},
			},
			expectError: false,
		},
		{
			name: "non-map dependencies",
			stateData: deployer.OpaqueState{
				"interopDepSet": map[string]interface{}{
					"dependencies": "not a map",
				},
			},
			expected: &config.Interop{
				Dependencies: map[string]config.StaticConfigDependency{},
			},
			expectError: false,
		},
		{
			name: "interopDepSet not a map",
			stateData: deployer.OpaqueState{
				"interopDepSet": "not a map",
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractInteropDepSet(tt.stateData)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
