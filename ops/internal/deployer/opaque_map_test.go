package deployer

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestReadOpcmImplSupportsV2State(t *testing.T) {
	expected := common.HexToAddress("0x80dcec9d21ce25b895ed26ca4ac8feb88c159eec")
	state := OpaqueState{
		"implementationsDeployment": map[string]any{
			"OpcmV2Impl": expected.Hex(),
		},
	}

	actual, err := state.ReadOpcmImpl()
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}
