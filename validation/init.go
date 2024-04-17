package validation

import (
	"embed"
	"fmt"
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

type StandardConfigTy struct {
	ResourceConfig ResourceConfig `toml:"resource_config"`
}

//go:embed standard-config.toml
var standardConfigFile embed.FS
var StandardConfig StandardConfigTy

func init() {
	data, err := fs.ReadFile(standardConfigFile, "standard-config.toml")
	if err != nil {
		panic(fmt.Errorf("error reading embedded file: %w", err))
	}

	err = toml.Unmarshal(data, &StandardConfig)
	if err != nil {
		panic(fmt.Errorf("error parsing embedded file: %w", err))
	}
}
