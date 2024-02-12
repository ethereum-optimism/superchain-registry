package superchain

import (
	"math/big"
)

var uint128Max, ok = big.NewInt(0).SetString("ffffffffffffffffffffffffffffffff", 16)

func init() {
	if !ok {
		panic("cannot construct uint128Max")
	}
}

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
	SubmissionInterval:        big.NewInt(120),
	L2BlockTime:               big.NewInt(2),
	FinalizationPeriodSeconds: big.NewInt(12),
}

// OPGoerliL2OOParams describes the L2OutputOracle parameters from OP Goerli
var OPGoerliL2OOParams = L2OOParams{
	SubmissionInterval:        big.NewInt(120),
	L2BlockTime:               big.NewInt(2),
	FinalizationPeriodSeconds: big.NewInt(12),
}

// OPGoerliDev0L2OOParams describes the L2OutputOracle parameters from OP Goerli
var OPGoerliDev0L2OOParams = L2OOParams{
	SubmissionInterval:        big.NewInt(120),
	L2BlockTime:               big.NewInt(2),
	FinalizationPeriodSeconds: big.NewInt(12),
}

// OPSepoliaL2OOParams describes the L2OutputOracle parameters from OP Goerli
var OPSepoliaL2OOParams = L2OOParams{
	SubmissionInterval:        big.NewInt(120),
	L2BlockTime:               big.NewInt(2),
	FinalizationPeriodSeconds: big.NewInt(12),
}

// OPSepoliaDev0L2OOParams describes the L2OutputOracle parameters from OP Goerli
var OPSepoliaDev0L2OOParams = L2OOParams{
	SubmissionInterval:        big.NewInt(120),
	L2BlockTime:               big.NewInt(2),
	FinalizationPeriodSeconds: big.NewInt(12),
}

type BigIntAndTolerance struct {
	Value  *big.Int
	Bounds [2]*big.Int
}

type Uint32AndTolerance struct {
	Value  uint32
	Bounds [2]uint32
}

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

type PreEcotoneGasPriceOracleParamsWithBounds struct {
	Decimals BigIntAndTolerance
	Overhead BigIntAndTolerance
	Scalar   BigIntAndTolerance
}

type EcotoneGasPriceOracleParamsWithBounds struct {
	Decimals          BigIntAndTolerance
	BlobBaseFeeScalar Uint32AndTolerance
	BaseFeeScalar     Uint32AndTolerance
}

type UpgradeFilter struct {
	PreEcotone PreEcotoneGasPriceOracleParamsWithBounds
	Ecotone    EcotoneGasPriceOracleParamsWithBounds
}

var GasPriceOracleParams = map[string]UpgradeFilter{
	"mainnet": {
		PreEcotone: PreEcotoneGasPriceOracleParamsWithBounds{
			Decimals: BigIntAndTolerance{big.NewInt(6), [2]*big.Int{big.NewInt(6), big.NewInt(6)}},
			Overhead: BigIntAndTolerance{big.NewInt(188), [2]*big.Int{big.NewInt(188), big.NewInt(188)}},
			Scalar:   BigIntAndTolerance{big.NewInt(684_000), [2]*big.Int{big.NewInt(684_000), big.NewInt(684_000)}},
		},
	},
	"sepolia": {
		PreEcotone: PreEcotoneGasPriceOracleParamsWithBounds{
			Decimals: BigIntAndTolerance{big.NewInt(6), [2]*big.Int{big.NewInt(6), big.NewInt(6)}},
			Overhead: BigIntAndTolerance{big.NewInt(188), [2]*big.Int{big.NewInt(188), big.NewInt(188)}},
			Scalar:   BigIntAndTolerance{big.NewInt(684_000), [2]*big.Int{big.NewInt(684_000), big.NewInt(684_000)}},
		},
	},
	"goerli": {
		PreEcotone: PreEcotoneGasPriceOracleParamsWithBounds{
			Decimals: BigIntAndTolerance{big.NewInt(6), [2]*big.Int{big.NewInt(6), big.NewInt(6)}},
			Overhead: BigIntAndTolerance{big.NewInt(2100), [2]*big.Int{big.NewInt(2100), big.NewInt(2100)}},
			Scalar:   BigIntAndTolerance{big.NewInt(100_000), [2]*big.Int{big.NewInt(100_000), big.NewInt(100_000)}},
		},
		Ecotone: EcotoneGasPriceOracleParamsWithBounds{
			Decimals:          BigIntAndTolerance{big.NewInt(6), [2]*big.Int{big.NewInt(6), big.NewInt(6)}},
			BlobBaseFeeScalar: Uint32AndTolerance{862_000, [2]uint32{862_000, 862_000}},
			BaseFeeScalar:     Uint32AndTolerance{7600, [2]uint32{7600, 7600}},
		},
	},
}
