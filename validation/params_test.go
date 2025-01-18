package validation

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

func TestRange(t *testing.T) {
	t.Run("within range", func(t *testing.T) {
		r := Range{1, 10}
		require.True(t, r.WithinRange(5))
	})

	t.Run("at lower bound", func(t *testing.T) {
		r := Range{1, 10}
		require.True(t, r.WithinRange(1))
	})

	t.Run("at upper bound", func(t *testing.T) {
		r := Range{1, 10}
		require.True(t, r.WithinRange(10))
	})

	t.Run("below range", func(t *testing.T) {
		r := Range{1, 10}
		require.False(t, r.WithinRange(0))
	})

	t.Run("above range", func(t *testing.T) {
		r := Range{1, 10}
		require.False(t, r.WithinRange(11))
	})

	t.Run("negative range", func(t *testing.T) {
		r := Range{-10, -1}
		require.True(t, r.WithinRange(-5))
	})

	t.Run("zero range", func(t *testing.T) {
		r := Range{0, 0}
		require.True(t, r.WithinRange(0))
	})
}

func TestConfigParams(t *testing.T) {
	t.Run("valid config unmarshal", func(t *testing.T) {
		tomlData := `
[rollup_config]
seq_window_size = [10, 20]
block_time = [1, 5]

[optimism_portal_2]
proof_maturity_delay_seconds = [100, 200]
dispute_game_finality_delay_seconds = [300, 400]
respected_game_type = 1

[resource_config]
max_resource_limit = 1000000
elasticity_multiplier = 10
base_fee_max_change_denominator = 8
minimum_base_fee = 1000
system_tx_max_gas = 1000000
maximum_base_fee = "1000000000"

[gas_price_oracle.pre-ecotone]
decimals = [6, 6]
overhead = [100, 200]
scalar = [1, 2]

[gas_price_oracle.ecotone]
decimals = [6, 6]
blob_base_fee_scalar = [1, 2]
base_fee_scalar = [1, 2]

[system_config]
gas_limit = [5000000, 10000000]

[proofs.permissioned]
game_type = 0
max_game_depth = 30
split_depth = 10
max_clock_duration = 3600
clock_extension = 300

[proofs.permissionless]
game_type = 1
max_game_depth = 30
split_depth = 10
max_clock_duration = 3600
clock_extension = 300
`
		var params ConfigParams
		err := toml.Unmarshal([]byte(tomlData), &params)
		require.NoError(t, err)

		// Test RollupConfig
		require.Equal(t, Range{10, 20}, params.RollupConfig.SeqWindowSize)
		require.Equal(t, Range{1, 5}, params.RollupConfig.BlockTime)

		// Test OptimismPortal2
		require.Equal(t, Range{100, 200}, params.OptimismPortal2.ProofMaturityDelaySeconds)
		require.Equal(t, Range{300, 400}, params.OptimismPortal2.DisputeGameFinalityDelaySeconds)
		require.Equal(t, 1, params.OptimismPortal2.RespectedGameType)

		// Test ResourceConfig
		require.Equal(t, 1000000, params.ResourceConfig.MaxResourceLimit)
		require.Equal(t, 10, params.ResourceConfig.ElasticityMultiplier)
		require.Equal(t, 8, params.ResourceConfig.BaseFeeMaxChangeDenominator)
		require.Equal(t, 1000, params.ResourceConfig.MinimumBaseFee)
		require.Equal(t, 1000000, params.ResourceConfig.SystemTxMaxGas)
		require.Equal(t, "1000000000", params.ResourceConfig.MaximumBaseFee)

		// Test GasPriceOracle
		require.Equal(t, Range{6, 6}, params.GasPriceOracle.PreEcotone.Decimals)
		require.Equal(t, Range{100, 200}, params.GasPriceOracle.PreEcotone.Overhead)
		require.Equal(t, Range{1, 2}, params.GasPriceOracle.PreEcotone.Scalar)

		require.Equal(t, Range{6, 6}, params.GasPriceOracle.Ecotone.Decimals)
		require.Equal(t, Range{1, 2}, params.GasPriceOracle.Ecotone.BlobBaseFeeScalar)
		require.Equal(t, Range{1, 2}, params.GasPriceOracle.Ecotone.BaseFeeScalar)

		// Test SystemConfig
		require.Equal(t, Range{5000000, 10000000}, params.SystemConfig.GasLimit)

		// Test Proofs
		require.Equal(t, uint32(0), params.Proofs.Permissioned.GameType)
		require.Equal(t, uint64(30), params.Proofs.Permissioned.MaxGameDepth)
		require.Equal(t, uint64(10), params.Proofs.Permissioned.SplitDepth)
		require.Equal(t, uint64(3600), params.Proofs.Permissioned.MaxClockDuration)
		require.Equal(t, uint64(300), params.Proofs.Permissioned.ClockExtension)

		require.Equal(t, uint32(1), params.Proofs.Permissionless.GameType)
		require.Equal(t, uint64(30), params.Proofs.Permissionless.MaxGameDepth)
		require.Equal(t, uint64(10), params.Proofs.Permissionless.SplitDepth)
		require.Equal(t, uint64(3600), params.Proofs.Permissionless.MaxClockDuration)
		require.Equal(t, uint64(300), params.Proofs.Permissionless.ClockExtension)
	})

	t.Run("invalid range format", func(t *testing.T) {
		tomlData := `
[rollup_config]
seq_window_size = [10]  # Invalid range format
block_time = [1, 5]
`
		var params ConfigParams
		err := toml.Unmarshal([]byte(tomlData), &params)
		require.Error(t, err)
	})

	t.Run("invalid number type", func(t *testing.T) {
		tomlData := `
[resource_config]
max_resource_limit = "not a number"
`
		var params ConfigParams
		err := toml.Unmarshal([]byte(tomlData), &params)
		require.Error(t, err)
	})

	t.Run("invalid range values", func(t *testing.T) {
		tomlData := `
[rollup_config]
seq_window_size = ["a", "b"]
block_time = [1, 5]
`
		var params ConfigParams
		err := toml.Unmarshal([]byte(tomlData), &params)
		require.Error(t, err)
	})

	t.Run("invalid uint values", func(t *testing.T) {
		tomlData := `
[proofs.permissioned]
game_type = -1  # Invalid negative value for uint
`
		var params ConfigParams
		err := toml.Unmarshal([]byte(tomlData), &params)
		require.Error(t, err)
		require.Contains(t, err.Error(), "-1 is out of range")
	})
}

func TestStandardConfigParams(t *testing.T) {
	t.Run("mainnet config loaded", func(t *testing.T) {
		require.NotEmpty(t, StandardConfigParamsMainnet)
		require.True(t, StandardConfigParamsMainnet.RollupConfig.BlockTime.WithinRange(2))
		require.True(t, StandardConfigParamsMainnet.SystemConfig.GasLimit.WithinRange(30000000))
	})

	t.Run("sepolia config loaded", func(t *testing.T) {
		require.NotEmpty(t, StandardConfigParamsSepolia)
		require.True(t, StandardConfigParamsSepolia.RollupConfig.BlockTime.WithinRange(2))
		require.True(t, StandardConfigParamsSepolia.SystemConfig.GasLimit.WithinRange(30000000))
	})
}
