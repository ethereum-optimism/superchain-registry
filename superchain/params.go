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
type GasPriceOracleParams struct {
type BedrockGasPriceOracleParams struct {
	Decimals *big.Int
	Overhead *big.Int
	Scalar   *big.Int
}

var OPMainnetBedrockGasPriceOracleParams = BedrockGasPriceOracleParams{
	Decimals: big.NewInt(6),
	Overhead: big.NewInt(188),
	Scalar:   big.NewInt(684000),
}

var OPGoerliBedrockGasPriceOracleParams = BedrockGasPriceOracleParams{
	Decimals: big.NewInt(6),
	Overhead: big.NewInt(2100),
	Scalar:   big.NewInt(1000000),
}

var OPSepoliaBedrockGasPriceOracleParams = BedrockGasPriceOracleParams{
	Decimals: big.NewInt(6),
	Overhead: big.NewInt(188),
	Scalar:   big.NewInt(684000),
}

type EcotoneGasPriceOracleParams struct {
	Decimals          *big.Int
	BlobBaseFeeScalar uint32
	BaseFeeScalar     uint32
}

var OPMainnetEcotoneGasPriceOracleParams = EcotoneGasPriceOracleParams{
	Decimals:          big.NewInt(6),
	BlobBaseFeeScalar: 0,
	BaseFeeScalar:     0,
}

var OPGoerliEcotoneGasPriceOracleParams = EcotoneGasPriceOracleParams{
	Decimals:          big.NewInt(6),
	BlobBaseFeeScalar: 0,       // TODO this parameter will update to 862000 very soon, see https://github.com/ethereum-optimism/superchain-ops/pull/58
	BaseFeeScalar:     1000000, // TODO this parameter will update to 7600 very soons, see https://github.com/ethereum-optimism/superchain-ops/pull/58
}

var OPSepoliaEcotoneGasPriceOracleParams = EcotoneGasPriceOracleParams{
	Decimals:          big.NewInt(6),
	BlobBaseFeeScalar: 0,
	BaseFeeScalar:     0,
}
