package manage

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/artifacts"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/common"
)

func InflateChainConfig(opd *deployer.OpDeployer, st deployer.OpaqueState, statePath string, idx int) (*config.StagedChain, error) {
	chainId, err := st.ReadL2ChainId(idx)
	if err != nil {
		return nil, fmt.Errorf("failed to read chain ID: %w", err)
	}

	rollup, err := opd.InspectRollup(statePath, chainId)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect rollup: %w", err)
	}

	dc, err := opd.InspectDeployConfig(statePath, chainId)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect deploy config: %w", err)
	}

	l1Contracts, err := st.ReadL1ContractsLocator()
	if err != nil {
		return nil, fmt.Errorf("failed to read L1 contracts locator: %w", err)
	}
	l1ContractsLocator, err := artifacts.NewLocatorFromURL(l1Contracts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse L1 contracts locator: %w", err)
	}

	l2contracts, err := st.ReadL2ContractsLocator()
	if err != nil {
		return nil, fmt.Errorf("failed to read L2 contracts locator: %w", err)
	}
	l2contractsLocator, err := artifacts.NewLocatorFromURL(l2contracts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse L2 contracts locator: %w", err)
	}

	cfg := new(config.StagedChain)

	cfg.ChainID = uint64(common.HexToHash(chainId).Big().Int64())
	cfg.BatchInboxAddr = config.NewChecksummedAddress(dc.BatchInboxAddress)
	cfg.BlockTime = dc.L2BlockTime
	cfg.SeqWindowSize = dc.SequencerWindowSize
	cfg.MaxSequencerDrift = dc.MaxSequencerDrift
	cfg.DataAvailabilityType = "eth-da"
	cfg.DeploymentL1ContractsVersion = l1ContractsLocator
	cfg.DeploymentL2ContractsVersion = l2contractsLocator
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

	cfg.Genesis = config.Genesis{
		L2Time: rollup.Genesis.L2Time,
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

	systemConfigProxy, err := st.ReadSystemConfigProxy(idx)
	if err != nil {
		return nil, fmt.Errorf("failed to read SystemConfigProxy: %w", err)
	}
	l1StandardBridgeProxy, err := st.ReadL1StandardBridgeProxy(idx)
	if err != nil {
		return nil, fmt.Errorf("failed to read L1StandardBridgeProxy: %w", err)
	}

	cfg.Addresses = config.Addresses{
		SystemConfigProxy:     config.NewChecksummedAddress(systemConfigProxy),
		L1StandardBridgeProxy: config.NewChecksummedAddress(l1StandardBridgeProxy),
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

func InflateSuperchainDefinition(name string, st deployer.OpaqueState) (*config.SuperchainDefinition, error) {
	protocolVersionsProxyAddress, err := st.ReadProtocolVersionsProxy()
	if err != nil {
		return nil, fmt.Errorf("failed to read protocol versions proxy address: %w", err)
	}
	superchainConfigProxyAddress, err := st.ReadSuperchainConfigProxy()
	if err != nil {
		return nil, fmt.Errorf("failed to read superchain config proxy address: %w", err)
	}
	opcmAddress, err := st.ReadOpcmAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to read opcm address: %w", err)
	}

	l1ChainID, err := st.ReadL1ChainID()
	if err != nil {
		return nil, fmt.Errorf("failed to read l1 chain id: %w", err)
	}

	sD := config.SuperchainDefinition{
		Name:                   name,
		ProtocolVersionsAddr:   config.NewChecksummedAddress(protocolVersionsProxyAddress),
		SuperchainConfigAddr:   config.NewChecksummedAddress(superchainConfigProxyAddress),
		OPContractsManagerAddr: config.NewChecksummedAddress(opcmAddress),
		Hardforks:              config.Hardforks{}, // superchain wide hardforks are added after chains are in the registry.
		L1: config.SuperchainL1{
			ChainID: l1ChainID,
		},
	}

	return &sD, nil
}

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
