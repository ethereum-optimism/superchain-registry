package superchain

import (
	"math/big"
)

type BigIntBounds [2]*big.Int

type Uint32Bounds [2]uint32

type PreEcotoneGasPriceOracleParams struct {
	Decimals *big.Int
	Overhead *big.Int
	Scalar   *big.Int
}

type EcotoneGasPriceOracleParams struct {
	Decimals          *big.Int
	BlobBaseFeeScalar uint32
	BaseFeeScalar     uint32
}

type PreEcotoneGasPriceOracleBounds struct {
	Decimals BigIntBounds
	Overhead BigIntBounds
	Scalar   BigIntBounds
}

type EcotoneGasPriceOracleBounds struct {
	Decimals          BigIntBounds
	BlobBaseFeeScalar Uint32Bounds
	BaseFeeScalar     Uint32Bounds
}

type UpgradeFilter struct {
	PreEcotone *PreEcotoneGasPriceOracleBounds
	Ecotone    *EcotoneGasPriceOracleBounds
}

func makeBigIntBounds(bounds [2]int64) BigIntBounds {
	return BigIntBounds([2]*big.Int{big.NewInt(bounds[0]), big.NewInt(bounds[1])})
}

var GasPriceOracleParams = map[string]UpgradeFilter{
	"mainnet": {
		PreEcotone: &PreEcotoneGasPriceOracleBounds{
			Decimals: makeBigIntBounds([2]int64{6, 6}),
			Overhead: makeBigIntBounds([2]int64{188, 188}),
			Scalar:   makeBigIntBounds([2]int64{684_000, 684_000}),
		},
		Ecotone: &EcotoneGasPriceOracleBounds{
			Decimals:          makeBigIntBounds([2]int64{6, 6}),
			BlobBaseFeeScalar: [2]uint32{0, 1e7},
			BaseFeeScalar:     [2]uint32{0, 1e7},
		},
	},
	"sepolia": {
		PreEcotone: &PreEcotoneGasPriceOracleBounds{
			Decimals: makeBigIntBounds([2]int64{6, 6}),
			Overhead: makeBigIntBounds([2]int64{188, 2_100}),
			Scalar:   makeBigIntBounds([2]int64{684_000, 1_000_000}),
		},
		Ecotone: &EcotoneGasPriceOracleBounds{
			Decimals:          makeBigIntBounds([2]int64{6, 6}),
			BlobBaseFeeScalar: [2]uint32{0, 1e7},
			BaseFeeScalar:     [2]uint32{0, 1e7},
		},
	},
}
