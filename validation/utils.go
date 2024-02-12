package validation

import "math/big"

// Returns true if |actual-desired) is within tolerance of desired. Tolerance is treated as a percentage.
var areCloseBigInts = func(actual, desired *big.Int, tolerance uint32) bool {
	difference := new(big.Int).Sub(desired, actual)                            // d - a
	difference100 := new(big.Int).Mul(big.NewInt(100), difference)             // 100(d - a)
	scaledTolerance := new(big.Int).Mul(desired, big.NewInt(int64(tolerance))) // dt
	return difference100.CmpAbs(scaledTolerance) <= 0                          // 100|d-a| <= dt implying |d-a|/d <= t/100
}

// Returns true if |actual-desired) is within tolerance of desired. Tolerance is treated as a percentage.
var areCloseInts = func(actual, desired uint32, tolerance uint32) bool {
	return areCloseBigInts(big.NewInt(int64(actual)), big.NewInt(int64(desired)), tolerance)

}
