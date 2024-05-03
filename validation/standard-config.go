package validation

import "math/big"

// StandardConfig is keyed by superchain target, e.g. "mainnet" or "sepolia" or "sepolia-dev-0"
var StandardConfig map[string]*StandardConfigTy

type ResourceConfig struct {
	MaxResourceLimit            uint32   `toml:"max_resource_limit"`
	ElasticityMultiplier        uint8    `toml:"elasticity_multiplier"`
	BaseFeeMaxChangeDenominator uint8    `toml:"base_fee_max_change_denominator"`
	MinimumBaseFee              uint32   `toml:"minimum_base_fee"`
	SystemTxMaxGas              uint32   `toml:"system_tx_max_gas"`
	MaximumBaseFee              *big.Int `toml:"maximum_base_fee"`
}

type L2OOParamsBounds struct {
	SubmissionInterval     BigIntBounds `toml:"submission_interval"`      // Interval in blocks at which checkpoints must be submitted.
	L2BlockTime            BigIntBounds `toml:"l2_block_time"`            // The time per L2 block, in seconds.
	ChallengePeriodSeconds BigIntBounds `toml:"challenge_period_seconds"` // Length of time for which an output root can be removed, and for which it is not considered finalized.
}

type GasPriceOracleBounds struct {
	PreEcotone PreEcotoneGasPriceOracleBounds `toml:"pre-ecotone"`
	Ecotone    EcotoneGasPriceOracleBounds    `toml:"ecotone"`
}

type BigIntBounds = [2]*big.Int

type (
	Uint32Bounds                   = [2]uint32
	PreEcotoneGasPriceOracleBounds struct {
		Decimals BigIntBounds `toml:"decimals"`
		Overhead BigIntBounds `toml:"overhead"`
		Scalar   BigIntBounds `toml:"scalar"`
	}
)

type EcotoneGasPriceOracleBounds struct {
	Decimals          BigIntBounds `toml:"decimals"`
	BlobBaseFeeScalar Uint32Bounds `toml:"blob_base_fee_scalar"`
	BaseFeeScalar     Uint32Bounds `toml:"base_fee_scalar"`
}

type StandardConfigTy struct {
	ResourceConfig ResourceConfig       `toml:"resource_config"`
	L2OOParams     L2OOParamsBounds     `toml:"l2_output_oracle"`
	GPOParams      GasPriceOracleBounds `toml:"gas_price_oracle"`
}
