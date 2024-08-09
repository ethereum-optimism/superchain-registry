package standard

import (
	"math/big"

	"github.com/ethereum-optimism/superchain-registry/superchain"
)

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

type RollupConfigBounds struct {
	AltDA               *superchain.AltDAConfig `toml:"alt_da"`
	BlockTime           [2]uint64               `toml:"block_time"`
	SequencerWindowSize [2]uint64               `toml:"seq_window_size"`
}

type OptimismPortal2Bounds struct {
	ProofMaturityDelaySeconds       [2]uint64 `toml:"proof_maturity_delay_seconds"`
	DisputeGameFinalityDelaySeconds [2]uint64 `toml:"dispute_game_finality_delay_seconds"`
	RespectedGameType               uint32    `toml:"respected_game_type"`
}

type Params struct {
	RollupConfig          RollupConfigBounds    `toml:"rollup_config"`
	OptimismPortal2Config OptimismPortal2Bounds `toml:"optimism_portal_2"`
	ResourceConfig        ResourceConfig        `toml:"resource_config"`
	GPOParams             GasPriceOracleBounds  `toml:"gas_price_oracle"`
	SystemConfig          SystemConfig          `toml:"system_config"`
}
