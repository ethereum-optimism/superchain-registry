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
	StartingBlockNumber       *big.Int
	StartingTimestamp         *big.Int
	SubmissionInterval        *big.Int
	L2BlockTime               *big.Int
	FinalizationPeriodSeconds *big.Int
}

// OPMainnetL2OOParams describes the L2OutputOracle parameters from OP Mainnet
var OPMainnetL2OOParams = L2OOParams{
	StartingBlockNumber:       big.NewInt(0),
	StartingTimestamp:         big.NewInt(1690493568),
	SubmissionInterval:        big.NewInt(120),
	L2BlockTime:               big.NewInt(2),
	FinalizationPeriodSeconds: big.NewInt(12),
}
