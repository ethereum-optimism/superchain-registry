package deployer

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeState(t *testing.T) {
	tests := []struct {
		Version int
		Merger  func(state OpaqueState) (OpaqueMap, OpaqueState, error)
	}{
		{1, MergeStateV1},
		{2, MergeStateV2},
		{3, MergeStateV3},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("v%d", tt.Version), func(t *testing.T) {
			input, err := ReadOpaqueStateFile(fmt.Sprintf("testdata/v%d-state-input.json", tt.Version))
			require.NoError(t, err)

			expectedState, err := os.ReadFile(fmt.Sprintf("testdata/v%d-state-output.json", tt.Version))
			require.NoError(t, err)
			expectedIntent, err := os.ReadFile(fmt.Sprintf("testdata/v%d-intent-output.json", tt.Version))
			require.NoError(t, err)

			mergedIntent, mergedState, err := tt.Merger(input)
			require.NoError(t, err)

			mergedStateJSON, err := json.Marshal(mergedState)
			require.NoError(t, err)
			mergedIntentJSON, err := json.Marshal(mergedIntent)
			require.NoError(t, err)

			require.JSONEq(t, string(expectedState), string(mergedStateJSON), "expected state invalid")
			require.JSONEq(t, string(expectedIntent), string(mergedIntentJSON), "expected intent invalid")
		})
	}
}

func TestReadOpaqueStateAcceptsEmptyLegacyAltDA(t *testing.T) {
	for _, filename := range []string{
		"testdata/v1-state-input.json",
		"testdata/v2-state-input.json",
		"testdata/v3-state-input.json",
		"configs/v4-state.json",
	} {
		t.Run(filename, func(t *testing.T) {
			_, err := ReadOpaqueStateFile(filename)
			require.NoError(t, err)
		})
	}
}

func TestRejectLegacyAltDAState(t *testing.T) {
	const address = "0x1111111111111111111111111111111111111111"
	tests := []struct {
		name  string
		state string
	}{
		{
			name:  "enabled config",
			state: `{"appliedIntent":{"chains":[{"dangerousAltDAConfig":{"useAltDA":true}}]}}`,
		},
		{
			name:  "disabled config with commitment",
			state: `{"appliedIntent":{"chains":[{"dangerousAltDAConfig":{"useAltDA":false,"daCommitmentType":"KeccakCommitment"}}]}}`,
		},
		{
			name:  "config on later chain",
			state: `{"appliedIntent":{"chains":[{}, {"dangerousAltDAConfig":{"daChallengeWindow":1}}]}}`,
		},
		{
			name:  "legacy proxy address",
			state: `{"opChainDeployments":[{"dataAvailabilityChallengeProxyAddress":"` + address + `"}]}`,
		},
		{
			name:  "legacy implementation address",
			state: `{"opChainDeployments":[{"dataAvailabilityChallengeImplAddress":"` + address + `"}]}`,
		},
		{
			name:  "v4 proxy address",
			state: `{"opChainDeployments":[{"AltDAChallengeProxy":"` + address + `"}]}`,
		},
		{
			name:  "v4 implementation address on later chain",
			state: `{"opChainDeployments":[{}, {"AltDAChallengeImpl":"` + address + `"}]}`,
		},
		{
			name:  "malformed address",
			state: `{"opChainDeployments":[{"AltDAChallengeProxy":"not-an-address"}]}`,
		},
		{
			name:  "empty address",
			state: `{"opChainDeployments":[{"AltDAChallengeProxy":""}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rejectLegacyAltDAState([]byte(tt.state))
			require.ErrorIs(t, err, ErrUnsupportedDataAvailability)
		})
	}
}

func TestRejectLegacyAltDAStateAllowsAbsentAndZeroValues(t *testing.T) {
	const state = `{
		"appliedIntent": {"chains": [{"dangerousAltDAConfig": {
			"useAltDA": false,
			"daCommitmentType": "",
			"daChallengeWindow": 0,
			"daResolveWindow": 0,
			"daBondSize": 0,
			"daResolverRefundPercentage": 0
		}}]},
		"opChainDeployments": [{
			"dataAvailabilityChallengeProxyAddress": "0x0000000000000000000000000000000000000000",
			"dataAvailabilityChallengeImplAddress": "0x0000000000000000000000000000000000000000",
			"AltDAChallengeProxy": null,
			"AltDAChallengeImpl": "0x0000000000000000000000000000000000000000"
		}]
	}`

	require.NoError(t, rejectLegacyAltDAState([]byte(state)))
}
