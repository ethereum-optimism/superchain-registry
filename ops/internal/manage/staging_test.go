package manage

import (
	"encoding/json"
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
		name          string
		stateDataJSON string
		expected      *config.Interop
		expectError   bool
	}{
		{
			name: "valid interop dependencies",
			stateDataJSON: `{
				"interopDepSet": {
					"dependencies": {
						"123": {},
						"456": {}
					}
				}
			}`,
			expected: &config.Interop{
				Dependencies: map[string]config.StaticConfigDependency{
					"123": {},
					"456": {},
				},
			},
			expectError: false,
		},
		{
			name:          "missing interopDepSet",
			stateDataJSON: `{}`,
			expected:      nil,
			expectError:   false,
		},
		{
			name: "null interopDepSet",
			stateDataJSON: `{
				"interopDepSet": null
			}`,
			expected:    nil,
			expectError: false,
		},
		{
			name: "empty dependencies",
			stateDataJSON: `{
				"interopDepSet": {
					"dependencies": {}
				}
			}`,
			expected:    nil,
			expectError: false,
		},
		{
			name: "non-map dependencies",
			stateDataJSON: `{
				"interopDepSet": {
					"dependencies": "not a map"
				}
			}`,
			expected:    nil,
			expectError: true,
		},
		{
			name: "interopDepSet not a map",
			stateDataJSON: `{
				"interopDepSet": "not a map"
			}`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os := new(deployer.OpaqueState)
			err := json.Unmarshal([]byte(tt.stateDataJSON), os)
			require.NoError(t, err, "failed to unmarshal testdata")
			require.NotNil(t, os)
			result, err := ExtractInteropDepSet(*os)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
