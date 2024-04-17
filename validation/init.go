package validation

import (
	"embed"
	"io/fs"
	"math/big"

	"github.com/BurntSushi/toml"
)

type ResourceConfig struct {
	MaxResourceLimit            uint32   `toml:"max_resource_limit"`
	ElasticityMultiplier        uint8    `toml:"elasticity_multiplier"`
	BaseFeeMaxChangeDenominator uint8    `toml:"base_fee_max_change_denominator"`
	MinimumBaseFee              uint32   `toml:"minimum_base_fee"`
	SystemTxMaxGas              uint32   `toml:"system_tx_max_gas"`
	MaximumBaseFee              *big.Int `toml:"maximum_base_fee"`
}

type L2OOParams struct {
	SubmissionInterval        *big.Int `toml:"submission_interval"`         // Interval in blocks at which checkpoints must be submitted.
	L2BlockTime               *big.Int `toml:"l2_block_time"`               // The time per L2 block, in seconds.
	FinalizationPeriodSeconds *big.Int `toml:"finalization_period_seconds"` // The minimum time (in seconds) that must elapse before a withdrawal can be finalized.
}

type StandardConfigTy struct {
	ResourceConfig ResourceConfig `toml:"resource_config"`
	L2OOParams     L2OOParams     `toml:"l2_output_oracle"`
}

//go:embed standard-config-mainnet.toml standard-config-sepolia.toml
var standardConfigFile embed.FS
var StandardConfigMainnet StandardConfigTy
var StandardConfigSepolia StandardConfigTy

func init() {
	var err error
	err = decodeTOMLFileIntoConfig("standard-config-mainnet.toml", &StandardConfigMainnet)
	if err != nil {
		panic(err)
	}
	err = decodeTOMLFileIntoConfig("standard-config-sepolia.toml", &StandardConfigSepolia)
	if err != nil {
		panic(err)
	}

}

func decodeTOMLFileIntoConfig(filename string, config *StandardConfigTy) error {
	data, err := fs.ReadFile(standardConfigFile, filename)
	if err != nil {
		return err
	}
	return toml.Unmarshal(data, config)
}
