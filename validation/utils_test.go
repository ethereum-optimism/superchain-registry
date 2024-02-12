package validation

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

// Returns true if actual is within bounds around desired. The bounds are (tolerance * desired / 100) either side.
var areCloseBigInts = func(actual, desired *big.Int, tolerance uint32) bool {
	difference := new(big.Int).Sub(desired, actual)                            // d - a
	difference100 := new(big.Int).Mul(big.NewInt(100), difference)             // 100(d - a)
	scaledTolerance := new(big.Int).Mul(desired, big.NewInt(int64(tolerance))) // dt
	return difference100.CmpAbs(scaledTolerance) <= 0                          // 100|d-a| <= dt implying |d-a|/d <= t/100
}

// Returns true if actual is within bounds around desired. The bounds are (tolerance * desired / 100) either side.
var areCloseInts = func(actual, desired uint32, tolerance uint32) bool {
	return areCloseBigInts(big.NewInt(int64(actual)), big.NewInt(int64(desired)), tolerance)

}

func TestAreCloseInts(t *testing.T) {
	tt := []struct {
		desired     uint32
		actual      uint32
		tolerance   uint32
		expectation bool
	}{
		{50, 60, 20, true},
		{50, 40, 20, true},
		{50, 60, 9, false},
		{50, 5, 10, false},
		{5, 0, 100, true},
		{100, 119, 20, true},
	}

	for _, test := range tt {
		t.Run(fmt.Sprintf("%+v", test), func(t *testing.T) {
			result := areCloseInts(test.actual, test.desired, test.tolerance)
			require.Equal(t, result, test.expectation)
		})
	}
}
