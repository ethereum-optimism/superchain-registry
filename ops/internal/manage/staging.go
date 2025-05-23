package manage

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/common"
)

func InflateChainConfig(opd *deployer.OpDeployer, statePath, chainId string, idx int) (*config.StagedChain, error) {

	rollup, err := opd.InspectRollup(statePath, chainId)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect rollup: %w", err)
	}

	dc, err := opd.InspectDeployConfig(statePath, chainId)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect deploy config: %w", err)
	}

	cfg := new(config.StagedChain)

	cfg.ChainID = uint64(common.HexToHash(chainId).Big().Int64())
	cfg.BatchInboxAddr = config.NewChecksummedAddress(dc.BatchInboxAddress)
	cfg.BlockTime = dc.L2BlockTime
	cfg.SeqWindowSize = dc.SequencerWindowSize
	cfg.MaxSequencerDrift = dc.MaxSequencerDrift
	cfg.DataAvailabilityType = "eth-da"
	// cfg.DeploymentL1ContractsVersion = st.AppliedIntent.L1ContractsLocator // TODO
	// cfg.DeploymentL2ContractsVersion = st.AppliedIntent.L2ContractsLocator // TODO
	cfg.DeploymentTxHash = new(common.Hash)
	cfg.BaseFeeVaultRecipient = *config.NewChecksummedAddress(dc.BaseFeeVaultRecipient)
	cfg.L1FeeVaultRecipient = *config.NewChecksummedAddress(dc.L1FeeVaultRecipient)
	cfg.SequencerFeeVaultRecipient = *config.NewChecksummedAddress(dc.SequencerFeeVaultRecipient)

	if dc.CustomGasTokenAddress != (common.Address{}) {
		cfg.GasPayingToken = config.NewChecksummedAddress(dc.CustomGasTokenAddress)
	}

	if err := CopyDeployConfigHFTimes(&dc.UpgradeScheduleDeployConfig, &cfg.Hardforks); err != nil {
		return nil, fmt.Errorf("failed to copy deploy config hardfork times: %w", err)
	}

	cfg.Optimism = config.Optimism{
		EIP1559Elasticity:        dc.EIP1559Elasticity,
		EIP1559Denominator:       dc.EIP1559Denominator,
		EIP1559DenominatorCanyon: dc.EIP1559DenominatorCanyon,
	}

	if dc.UseAltDA {
		cfg.AltDA = &config.AltDA{
			DaChallengeContractAddress: config.ChecksummedAddress(dc.DAChallengeProxy),
			DaChallengeWindow:          dc.DAChallengeWindow,
			DaResolveWindow:            dc.DAResolveWindow,
			DaCommitmentType:           dc.DACommitmentType,
		}
		cfg.Addresses.DAChallengeAddress = config.NewChecksummedAddress(dc.DAChallengeProxy)
		cfg.DataAvailabilityType = "alt-da"
	}

	// chainState := st.Chains[0]
	cfg.Genesis = config.Genesis{
		// L2Time: uint64(chainState.StartBlock.Time),
		L1: config.GenesisRef{
			Hash:   rollup.Genesis.L1.Hash,
			Number: rollup.Genesis.L1.Number,
		},
		L2: config.GenesisRef{
			Hash:   rollup.Genesis.L2.Hash,
			Number: rollup.Genesis.L2.Number,
		},
		SystemConfig: config.SystemConfig{
			BatcherAddr: *config.NewChecksummedAddress(rollup.Genesis.SystemConfig.BatcherAddr),
			Overhead:    common.Hash(rollup.Genesis.SystemConfig.Overhead),
			Scalar:      common.Hash(rollup.Genesis.SystemConfig.Scalar),
			GasLimit:    rollup.Genesis.SystemConfig.GasLimit,
		},
	}

	cfg.Roles = config.Roles{
		SystemConfigOwner: config.NewChecksummedAddress(dc.FinalSystemOwner),
		ProxyAdminOwner:   config.NewChecksummedAddress(dc.ProxyAdminOwner),
		Guardian:          config.NewChecksummedAddress(dc.SuperchainConfigGuardian),
		Proposer:          config.NewChecksummedAddress(dc.L2OutputOracleProposer),
		UnsafeBlockSigner: config.NewChecksummedAddress(dc.P2PSequencerAddress),
		BatchSubmitter:    config.NewChecksummedAddress(dc.BatchSenderAddress),
		// Challenger:        config.NewChecksummedAddress(chainIntent.Roles.Challenger), // TODO
	}

	// cfg.Addresses = config.Addresses{ // TODO
	// 	AddressManager:                    config.NewChecksummedAddress(chainState.AddressManagerAddress),
	// 	L1CrossDomainMessengerProxy:       config.NewChecksummedAddress(chainState.L1CrossDomainMessengerProxyAddress),
	// 	L1ERC721BridgeProxy:               config.NewChecksummedAddress(chainState.L1ERC721BridgeProxyAddress),
	// 	L1StandardBridgeProxy:             config.NewChecksummedAddress(chainState.L1StandardBridgeProxyAddress),
	// 	OptimismMintableERC20FactoryProxy: config.NewChecksummedAddress(chainState.OptimismMintableERC20FactoryProxyAddress),
	// 	OptimismPortalProxy:               config.NewChecksummedAddress(chainState.OptimismPortalProxyAddress),
	// 	SystemConfigProxy:                 config.NewChecksummedAddress(chainState.SystemConfigProxyAddress),
	// 	ProxyAdmin:                        config.NewChecksummedAddress(chainState.ProxyAdminAddress),
	// 	SuperchainConfig:                  config.NewChecksummedAddress(st.SuperchainDeployment.SuperchainConfigProxyAddress),
	// 	AnchorStateRegistryProxy:          config.NewChecksummedAddress(chainState.AnchorStateRegistryProxyAddress),
	// 	DelayedWETHProxy:                  config.NewChecksummedAddress(chainState.DelayedWETHPermissionedGameProxyAddress),
	// 	DisputeGameFactoryProxy:           config.NewChecksummedAddress(chainState.DisputeGameFactoryProxyAddress),
	// 	PermissionedDisputeGame:           config.NewChecksummedAddress(chainState.PermissionedDisputeGameAddress),
	// }

	return cfg, nil
}

func CopyDeployConfigHFTimes(src *genesis.UpgradeScheduleDeployConfig, dst *config.Hardforks) error {
	if src == nil || dst == nil {
		return errors.New("source and destination must not be nil")
	}

	srcVal := reflect.ValueOf(src).Elem()
	srcType := srcVal.Type()
	dstVal := reflect.ValueOf(dst).Elem()

	// Iterate through source struct fields
	for i := 0; i < srcType.NumField(); i++ {
		field := srcType.Field(i)
		fieldName := field.Name

		// Only process fields that start with "L2Genesis" and end with "TimeOffset"
		// Also skip Regolith since it's not included in the configs.
		if !strings.HasPrefix(fieldName, "L2Genesis") ||
			!strings.HasSuffix(fieldName, "TimeOffset") ||
			strings.Contains(fieldName, "Regolith") {
			continue
		}

		// Extract the hardfork name (e.g., "Canyon" from "L2GenesisCanyonTimeOffset")
		hardforkName := strings.TrimSuffix(strings.TrimPrefix(fieldName, "L2Genesis"), "TimeOffset")

		// Construct the destination field name by adding "Time" suffix
		dstFieldName := hardforkName + "Time"

		// Get the destination field
		dstField := dstVal.FieldByName(dstFieldName)

		// Get the source field value
		srcField := srcVal.Field(i)
		if srcField.IsNil() {
			continue // Skip if source value is nil
		}

		if !dstField.IsValid() {
			return fmt.Errorf("destination field %s doesn't exist", dstFieldName)
		}

		// Create a new HardforkTime pointer and set its value
		newHardforkTime := new(config.HardforkTime)
		*newHardforkTime = config.HardforkTime(srcField.Elem().Uint())

		// Set the destination field
		dstField.Set(reflect.ValueOf(newHardforkTime))
	}

	return nil
}

var (
	ErrNoStagedConfig               = errors.New("no staged chain config found")
	ErrNoStagedSuperchainDefinition = errors.New("no staged superchain definition found")
)

func StagedChainConfigs(rootP string) ([]*config.StagedChain, error) {
	tomls, err := paths.CollectFiles(paths.StagingDir(rootP), paths.ChainConfigMatcher())
	if err != nil {
		return nil, fmt.Errorf("failed to collect staged chain configs: %w", err)
	}
	if len(tomls) == 0 {
		return nil, ErrNoStagedConfig
	}

	chainCfgs := make([]*config.StagedChain, len(tomls))
	for i, cfgFilename := range tomls {

		chainCfg := new(config.StagedChain)
		if err := paths.ReadTOMLFile(cfgFilename, chainCfg); err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", cfgFilename, err)
		}
		chainCfg.ShortName = strings.TrimSuffix(filepath.Base(cfgFilename), ".toml")
		chainCfgs[i] = chainCfg
	}
	return chainCfgs, nil
}

// StagedSuperchainDefinition finds a superchain.toml file in the staging directory
// (if it exists) and returns the parsed SuperchainDefinition struct.
func StagedSuperchainDefinition(rootP string) (*config.SuperchainDefinition, error) {
	// find the superchain.toml file
	files, err := paths.CollectFiles(paths.StagingDir(rootP), paths.SuperchainDefinitionMatcher())
	if err != nil {
		return nil, fmt.Errorf("failed to collect staged superchain definition: %w",
			err)
	}
	if len(files) == 0 {
		return nil, ErrNoStagedSuperchainDefinition
	}
	sM := new(config.SuperchainDefinition)
	err = paths.ReadTOMLFile(files[0], sM)

	return sM, err
}
