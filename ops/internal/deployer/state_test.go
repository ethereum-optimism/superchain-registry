package deployer

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/tomwright/dasel"
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

func TestMergeStateV4_1CopiesOperatorFeeVaultRecipient(t *testing.T) {
	const expected = "0xea9ecc2f4d32e075e961afffd522bbf88bba40a9"

	userIntent, err := StandardIntentV4_1(11155111)
	require.NoError(t, err)
	userIntentNode := dasel.New(userIntent)
	require.NoError(t, userIntentNode.Put("l1ChainID", int64(11155111)))
	mustPutString(userIntentNode, "chains.[0].operatorFeeVaultRecipient", common.HexToAddress(expected))

	userState, err := StandardStateV4(11155111)
	require.NoError(t, err)
	userState["appliedIntent"] = userIntent
	require.NoError(t, dasel.New(userState).Put("opChainDeployments.[0].startBlock", map[string]any{
		"hash":      common.Hash{}.Hex(),
		"number":    "0x0",
		"timestamp": "0x0",
	}))

	mergedIntent, _, err := MergeStateV4_1(userState)
	require.NoError(t, err)

	actual, err := QueryOpaqueMap[string](OpaqueState(mergedIntent), "chains.[0].operatorFeeVaultRecipient")
	require.NoError(t, err)
	require.Equal(t, common.HexToAddress(expected).Hex(), actual)
}
