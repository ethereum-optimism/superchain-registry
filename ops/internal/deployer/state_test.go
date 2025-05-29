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
		Merger  func(mapping OpaqueMapping) (OpaqueMapping, OpaqueMapping, error)
	}{
		{1, MergeStateV1},
		{2, MergeStateV2},
		{3, MergeStateV3},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("v%d", tt.Version), func(t *testing.T) {
			input, err := ReadOpaqueMappingFile(fmt.Sprintf("testdata/v%d-state-input.json", tt.Version))
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
