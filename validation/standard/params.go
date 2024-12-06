package standard

import (
	"fmt"
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

type (
	BigIntBounds                   [2]*big.Int
	Uint32Bounds                   [2]uint32
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

func (rc *ResourceConfig) Check() error {
	if rc.MaxResourceLimit == 0 {
		return fmt.Errorf("MaxResourceLimit is 0")
	}
	if rc.ElasticityMultiplier == 0 {
		return fmt.Errorf("ElasticityMultiplier is 0")
	}
	if rc.BaseFeeMaxChangeDenominator == 0 {
		return fmt.Errorf("BaseFeeMaxChangeDenominator is 0")
	}
	if rc.MinimumBaseFee == 0 {
		return fmt.Errorf("MinimumBaseFee is 0")
	}
	if rc.SystemTxMaxGas == 0 {
		return fmt.Errorf("SystemTxMaxGas is 0")
	}
	if rc.MaximumBaseFee == nil || rc.MaximumBaseFee.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("MaximumBaseFee is 0 or nil")
	}
	return nil
}

func (sc *SystemConfig) Check() error {
	if sc.GasLimit[1] == 0 {
		return fmt.Errorf("GasLimit upper bound is 0")
	}
	return nil
}

func (bb BigIntBounds) Check() error {
	if bb[1] == nil || bb[1].Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("BigIntBounds upper bound is 0 or nil")
	}
	return nil
}

func (ub Uint32Bounds) Check() error {
	if ub[1] == 0 {
		return fmt.Errorf("Uint32Bounds upper bound is 0")
	}
	return nil
}

func (pe *PreEcotoneGasPriceOracleBounds) Check() error {
	if err := pe.Decimals.Check(); err != nil {
		return fmt.Errorf("Decimals: %w", err)
	}
	if err := pe.Overhead.Check(); err != nil {
		return fmt.Errorf("Overhead: %w", err)
	}
	if err := pe.Scalar.Check(); err != nil {
		return fmt.Errorf("Scalar: %w", err)
	}
	return nil
}

func (ec *EcotoneGasPriceOracleBounds) Check() error {
	if err := ec.Decimals.Check(); err != nil {
		return fmt.Errorf("Decimals: %w", err)
	}
	if err := ec.BlobBaseFeeScalar.Check(); err != nil {
		return fmt.Errorf("BlobBaseFeeScalar: %w", err)
	}
	if err := ec.BaseFeeScalar.Check(); err != nil {
		return fmt.Errorf("BaseFeeScalar: %w", err)
	}
	return nil
}

func (rc *RollupConfigBounds) Check() error {
	if rc.BlockTime[1] == 0 {
		return fmt.Errorf("BlockTime upper bound is 0")
	}
	if rc.SequencerWindowSize[1] == 0 {
		return fmt.Errorf("SequencerWindowSize upper bound is 0")
	}
	return nil
}

func (op *OptimismPortal2Bounds) Check() error {
	if op.ProofMaturityDelaySeconds[1] == 0 {
		return fmt.Errorf("ProofMaturityDelaySeconds upper bound is 0")
	}
	if op.DisputeGameFinalityDelaySeconds[1] == 0 {
		return fmt.Errorf("DisputeGameFinalityDelaySeconds upper bound is 0")
	}
	return nil
}

func (p *Params) Check() error {
	if err := p.RollupConfig.Check(); err != nil {
		return fmt.Errorf("RollupConfig: %w", err)
	}
	if err := p.OptimismPortal2Config.Check(); err != nil {
		return fmt.Errorf("OptimismPortal2Config: %w", err)
	}
	if err := p.ResourceConfig.Check(); err != nil {
		return fmt.Errorf("ResourceConfig: %w", err)
	}
	if err := p.GPOParams.PreEcotone.Check(); err != nil {
		return fmt.Errorf("PreEcotone: %w", err)
	}
	if err := p.GPOParams.Ecotone.Check(); err != nil {
		return fmt.Errorf("Ecotone: %w", err)
	}
	if err := p.SystemConfig.Check(); err != nil {
		return fmt.Errorf("SystemConfig: %w", err)
	}
	return nil
}
