package superchain

import (
	"math/big"
)

var uint128Max = func() *big.Int {
	r, ok := new(big.Int).SetString("ffffffffffffffffffffffffffffffff", 16)
	if !ok {
		panic("cannot construct uint128Max")
	}
	return r
}()

type ResourceConfig struct {
	MaxResourceLimit            uint32
	ElasticityMultiplier        uint8
	BaseFeeMaxChangeDenominator uint8
	MinimumBaseFee              uint32
	SystemTxMaxGas              uint32
	MaximumBaseFee              *big.Int
}

// OPMainnetResourceConfig describes the resource metering configuration from OP Mainnet
var OPMainnetResourceConfig = ResourceConfig{
	MaxResourceLimit:            20000000,
	ElasticityMultiplier:        10,
	BaseFeeMaxChangeDenominator: 8,
	MinimumBaseFee:              1000000000,
	SystemTxMaxGas:              1000000,
	MaximumBaseFee:              uint128Max,
}

type L2OOParams struct {
	SubmissionInterval        *big.Int // Interval in blocks at which checkpoints must be submitted.
	L2BlockTime               *big.Int // The time per L2 block, in seconds.
	FinalizationPeriodSeconds *big.Int // The minimum time (in seconds) that must elapse before a withdrawal can be finalized.
}

// OPMainnetL2OOParams describes the L2OutputOracle parameters from OP Mainnet
var OPMainnetL2OOParams = L2OOParams{
	SubmissionInterval:        big.NewInt(1800),
	L2BlockTime:               big.NewInt(2),
	FinalizationPeriodSeconds: big.NewInt(604800),
}

// OPSepoliaL2OOParams describes the L2OutputOracle parameters from OP Sepolia
var OPSepoliaL2OOParams = L2OOParams{
	SubmissionInterval:        big.NewInt(120),
	L2BlockTime:               big.NewInt(2),
	FinalizationPeriodSeconds: big.NewInt(12),
}

// OPSepoliaDev0L2OOParams describes the L2OutputOracle parameters from OP Sepolia
var OPSepoliaDev0L2OOParams = L2OOParams{
	SubmissionInterval:        big.NewInt(120),
	L2BlockTime:               big.NewInt(2),
	FinalizationPeriodSeconds: big.NewInt(12),
}

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
