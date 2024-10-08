package validation

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isBigIntWithinBounds returns true if actual is within bounds, where the bounds are [lower bound, upper bound] and are inclusive.
var isBigIntWithinBounds = func(actual *big.Int, bounds [2]*big.Int) bool {
	if (bounds[1].Cmp(bounds[0])) < 0 {
		panic("bounds are in wrong order")
	}
	return (actual.Cmp(bounds[0]) >= 0 && actual.Cmp(bounds[1]) <= 0)
}

// isIntWithinBounds returns true if actual is within bounds, where the bounds are [lower bound, upper bound] and are inclusive.
func isIntWithinBounds[T uint32 | uint64](actual T, bounds [2]T) bool {
	if bounds[1] < bounds[0] {
		panic("bounds are in wrong order")
	}
	return (actual >= bounds[0] && actual <= bounds[1])
}

// assertInBounds fails the test (but not immediately) if the passed param is outside of the passed bounds.
func assertIntInBounds[T uint32 | uint64](t *testing.T, name string, got T, want [2]T) {
	assert.True(t,
		isIntWithinBounds(got, want),
		fmt.Sprintf("Incorrect %s, %d is not within bounds %d", name, got, want))
}

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
