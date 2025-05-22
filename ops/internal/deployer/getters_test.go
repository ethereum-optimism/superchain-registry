package deployer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetNumChains(t *testing.T) {
	tests := []struct {
		Version int
	}{
		{2},
		{3},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("v%d", tt.Version), func(t *testing.T) {
			input, err := ReadOpaqueMappingFile(fmt.Sprintf("testdata/v%d-state-input.json", tt.Version))
			require.NoError(t, err)

			numChains, err := GetNumChains(input)
			require.NoError(t, err)
			require.Equal(t, 1, numChains, "expected 1 chain in state file")
		})
	}

}
