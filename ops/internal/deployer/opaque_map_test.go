package deployer

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestReadOpcmImpl(t *testing.T) {
	const expected = "0x44e197058fb98fb3618453b3d90cdef6f5db8297"

	tests := []struct {
		name  string
		state OpaqueState
	}{
		{
			name: "legacy OpcmImpl",
			state: OpaqueState{
				"implementationsDeployment": map[string]any{"OpcmImpl": expected},
			},
		},
		{
			name: "v0.7 OpcmV2Impl",
			state: OpaqueState{
				"implementationsDeployment": map[string]any{"OpcmV2Impl": expected},
			},
		},
		{
			name: "legacy lowercase opcmAddress",
			state: OpaqueState{
				"implementationsDeployment": map[string]any{"opcmAddress": expected},
			},
		},
		{
			name: "intent opcmAddress fallback",
			state: OpaqueState{
				"appliedIntent": map[string]any{"opcmAddress": expected},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := tt.state.ReadOpcmImpl()
			require.NoError(t, err)
			require.Equal(t, common.HexToAddress(expected), actual)
		})
	}
}
