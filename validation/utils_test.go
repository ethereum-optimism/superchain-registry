package validation

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// perChainTestName ensures test can easily be filtered by chain name or chain id using the -run=regex testflag.
func perChainTestName(chain *superchain.ChainConfig) string {
	return chain.Name + fmt.Sprintf(" (%d)", chain.ChainID)
}

// isBigIntWithinBounds returns true if actual is within bounds, where the bounds are [lower bound, upper bound] and are inclusive.
var isBigIntWithinBounds = func(actual *big.Int, bounds [2]*big.Int) bool {
	if (bounds[1].Cmp(bounds[0])) < 0 {
		panic("bounds are in wrong order")
	}
	return (actual.Cmp(bounds[0]) >= 0 && actual.Cmp(bounds[1]) <= 0)
}

// isWithinBounds returns true if actual is within bounds, where the bounds are [lower bound, upper bound] and are inclusive.
func isWithinBounds[T uint32 | uint64](actual T, bounds [2]T) bool {
	if bounds[1] < bounds[0] {
		panic("bounds are in wrong order")
	}
	return (actual >= bounds[0] && actual <= bounds[1])
}

// assertInBounds fails the test (but not immediately) if the passed param is outside of the passed bounds.
var assertInBounds = func(t *testing.T, name string, got *big.Int, want [2]*big.Int) {
	assert.True(t,
		isBigIntWithinBounds(got, want),
		fmt.Sprintf("Incorrect %s, %d is not within bounds %d", name, got, want))
}

// assertInBounds fails the test (but not immediately) if the passed param is outside of the passed bounds.
func assertIntInBounds[T uint32 | uint64](t *testing.T, name string, got T, want [2]T) {
	assert.True(t,
		isWithinBounds(got, want),
		fmt.Sprintf("Incorrect %s, %d is not within bounds %d", name, got, want))
}

func TestAreCloseInts(t *testing.T) {
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
			result := isWithinBounds(test.actual, test.bounds)
			require.Equal(t, test.expectation, result)
		})
	}
}

func SkipCheckIfFrontierChain(t *testing.T, chain superchain.ChainConfig) {
	if chain.SuperchainLevel == superchain.Frontier {
		t.Skip()
	}
}
