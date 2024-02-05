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
	StartingBlockNumber       *big.Int // The number of the first L2 block.
	StartingTimestamp         *big.Int // The timestamp of the first L2 block.
	SubmissionInterval        *big.Int // Interval in blocks at which checkpoints must be submitted.
	L2BlockTime               *big.Int // The time per L2 block, in seconds.
	FinalizationPeriodSeconds *big.Int // The minimum time (in seconds) that must elapse before a withdrawal can be finalized.
}

// OPMainnetL2OOParams describes the L2OutputOracle parameters from OP Mainnet
var OPMainnetL2OOParams = L2OOParams{
	StartingBlockNumber:       big.NewInt(0),
	StartingTimestamp:         big.NewInt(1690493568),
	SubmissionInterval:        big.NewInt(120),
	L2BlockTime:               big.NewInt(2),
	FinalizationPeriodSeconds: big.NewInt(12),
}

// OPGoerliL2OOParams describes the L2OutputOracle parameters from OP Goerli
var OPGoerliL2OOParams = L2OOParams{
	StartingBlockNumber:       big.NewInt(4061224),
	StartingTimestamp:         big.NewInt(1673550516),
	SubmissionInterval:        big.NewInt(120),
	L2BlockTime:               big.NewInt(2),
	FinalizationPeriodSeconds: big.NewInt(12),
}

// OPGoerliDev0L2OOParams describes the L2OutputOracle parameters from OP Goerli
var OPGoerliDev0L2OOParams = L2OOParams{
	StartingBlockNumber:       big.NewInt(0),
	StartingTimestamp:         big.NewInt(1690493568),
	SubmissionInterval:        big.NewInt(120),
	L2BlockTime:               big.NewInt(2),
	FinalizationPeriodSeconds: big.NewInt(12),
}

// OPSepoliaL2OOParams describes the L2OutputOracle parameters from OP Goerli
var OPSepoliaL2OOParams = L2OOParams{
	StartingBlockNumber:       big.NewInt(0),
	StartingTimestamp:         big.NewInt(1690493568),
	SubmissionInterval:        big.NewInt(120),
	L2BlockTime:               big.NewInt(2),
	FinalizationPeriodSeconds: big.NewInt(12),
}

// OPSepoliaDev0L2OOParams describes the L2OutputOracle parameters from OP Goerli
var OPSepoliaDev0L2OOParams = L2OOParams{
	StartingBlockNumber:       big.NewInt(0),
	StartingTimestamp:         big.NewInt(1690493568),
	SubmissionInterval:        big.NewInt(120),
	L2BlockTime:               big.NewInt(2),
	FinalizationPeriodSeconds: big.NewInt(12),
}
