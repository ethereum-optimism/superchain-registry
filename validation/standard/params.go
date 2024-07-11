package standard

import "math/big"

type ResourceConfig struct {
	MaxResourceLimit            uint32   `toml:"max_resource_limit"`
	ElasticityMultiplier        uint8    `toml:"elasticity_multiplier"`
	BaseFeeMaxChangeDenominator uint8    `toml:"base_fee_max_change_denominator"`
	MinimumBaseFee              uint32   `toml:"minimum_base_fee"`
	SystemTxMaxGas              uint32   `toml:"system_tx_max_gas"`
	MaximumBaseFee              *big.Int `toml:"maximum_base_fee"`
}

type GasPriceOracleBounds struct {
	PreEcotone PreEcotoneGasPriceOracleBounds `toml:"pre-ecotone"`
	Ecotone    EcotoneGasPriceOracleBounds    `toml:"ecotone"`
}

type SystemConfig struct {
	GasLimit [2]uint64 `toml:"gas_limit"`
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
type Params struct {
	ResourceConfig ResourceConfig       `toml:"resource_config"`
	GPOParams      GasPriceOracleBounds `toml:"gas_price_oracle"`
	SystemConfig   SystemConfig         `toml:"system_config"`
}
