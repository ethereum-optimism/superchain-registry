package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsIntWithinBounds(t *testing.T) {
	tt := []struct {
		actual      uint32
		bounds      [2]uint32
		expectation bool
	}{
		{50, [2]uint32{50, 50}, true},
		{50, [2]uint32{40, 60}, true},
		{50, [2]uint32{40, 50}, true},
		{50, [2]uint32{50, 60}, true},
		{50, [2]uint32{50, 50}, true},
		{50, [2]uint32{30, 50}, true},
		{51, [2]uint32{30, 50}, false},
		{29, [2]uint32{30, 50}, false},
	}

	for _, test := range tt {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			result := isIntWithinBounds(test.actual, test.bounds)
			require.Equal(t, test.expectation, result)
		})
	}
}
