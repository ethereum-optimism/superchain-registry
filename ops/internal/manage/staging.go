package manage

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/inspect"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/backend/depset"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/common"
)

func InflateChainConfig(st *state.State, idx int) (*config.StagedChain, error) {
	if idx >= len(st.AppliedIntent.Chains) {
		return nil, errors.New("index out of bounds")
	}

	chainIntent := st.AppliedIntent.Chains[idx]
	chainID := chainIntent.ID
	dc, err := inspect.DeployConfig(st, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect deploy config: %w", err)
	}

	_, rollup, err := inspect.GenesisAndRollup(st, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect genesis and rollup: %w", err)
	}

	cfg := new(config.StagedChain)
	cfg.ChainID = chainID.Big().Uint64()
	cfg.BatchInboxAddr = config.NewChecksummedAddress(dc.BatchInboxAddress)
	cfg.BlockTime = dc.L2BlockTime
	cfg.SeqWindowSize = dc.SequencerWindowSize
	cfg.MaxSequencerDrift = dc.MaxSequencerDrift
	cfg.DataAvailabilityType = "eth-da"
	cfg.DeploymentL1ContractsVersion = st.AppliedIntent.L1ContractsLocator
	cfg.DeploymentL2ContractsVersion = st.AppliedIntent.L2ContractsLocator
	cfg.DeploymentTxHash = new(common.Hash)
	cfg.BaseFeeVaultRecipient = *config.NewChecksummedAddress(chainIntent.BaseFeeVaultRecipient)
	cfg.L1FeeVaultRecipient = *config.NewChecksummedAddress(chainIntent.L1FeeVaultRecipient)
	cfg.SequencerFeeVaultRecipient = *config.NewChecksummedAddress(chainIntent.SequencerFeeVaultRecipient)

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

	chainState := st.Chains[0]
	cfg.Genesis = config.Genesis{
		L2Time: uint64(chainState.StartBlock.Time),
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
		SystemConfigOwner: config.NewChecksummedAddress(chainIntent.Roles.SystemConfigOwner),
		ProxyAdminOwner:   config.NewChecksummedAddress(chainIntent.Roles.L1ProxyAdminOwner),
		Guardian:          config.NewChecksummedAddress(st.AppliedIntent.SuperchainRoles.SuperchainGuardian),
		Proposer:          config.NewChecksummedAddress(chainIntent.Roles.Proposer),
		UnsafeBlockSigner: config.NewChecksummedAddress(chainIntent.Roles.UnsafeBlockSigner),
		BatchSubmitter:    config.NewChecksummedAddress(chainIntent.Roles.Batcher),
		Challenger:        config.NewChecksummedAddress(chainIntent.Roles.Challenger),
	}

	cfg.Addresses = config.Addresses{
		AddressManager:                    config.NewChecksummedAddress(chainState.AddressManagerImpl),
		L1CrossDomainMessengerProxy:       config.NewChecksummedAddress(chainState.L1CrossDomainMessengerProxy),
		L1ERC721BridgeProxy:               config.NewChecksummedAddress(chainState.L1Erc721BridgeProxy),
		L1StandardBridgeProxy:             config.NewChecksummedAddress(chainState.L1StandardBridgeProxy),
		OptimismMintableERC20FactoryProxy: config.NewChecksummedAddress(chainState.OptimismMintableErc20FactoryProxy),
		OptimismPortalProxy:               config.NewChecksummedAddress(chainState.OptimismPortalProxy),
		SystemConfigProxy:                 config.NewChecksummedAddress(chainState.SystemConfigProxy),
		ProxyAdmin:                        config.NewChecksummedAddress(chainState.OpChainProxyAdminImpl),
		SuperchainConfig:                  config.NewChecksummedAddress(st.SuperchainDeployment.SuperchainConfigProxy),
		AnchorStateRegistryProxy:          config.NewChecksummedAddress(chainState.AnchorStateRegistryProxy),
		DelayedWETHProxy:                  config.NewChecksummedAddress(chainState.DelayedWethPermissionedGameProxy),
		DisputeGameFactoryProxy:           config.NewChecksummedAddress(chainState.DisputeGameFactoryProxy),
		PermissionedDisputeGame:           config.NewChecksummedAddress(chainState.PermissionedDisputeGameImpl),
	}

	// Dependency Set Inflation
	if isChainInDependencySet(chainID, st.InteropDepSet) {
		cfg.Interop = new(config.Interop)
		cfg.Interop.Dependencies = convertToSCRDepset(st.InteropDepSet)
	}

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

// isChainInDependencySet checks if a chain ID is in the supplied dependency set
func isChainInDependencySet(chainID common.Hash, depSet *depset.StaticConfigDependencySet) bool {
	if depSet == nil {
		return false
	}
	for id := range depSet.Dependencies() {
		if id.ToBig() == chainID.Big() {
			return true
		}
	}
	return false
}

// convertToSCRDepset converts a prestate.DependencySet to a config.DependencySet
func convertToSCRDepset(depSet *depset.StaticConfigDependencySet) map[string]config.Dependency {
	if depSet == nil {
		return nil
	}
	if depSet.Dependencies == nil {
		return nil
	}
	ds := make(map[string]config.Dependency, len(depSet.Dependencies()))
	for k, v := range depSet.Dependencies() {
		ds[k.String()] = config.Dependency{
			ChainIndex:     uint32(v.ChainIndex),
			ActivationTime: v.ActivationTime,
		}
	}
	return ds
}
