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

func InflateChainConfig(opd *deployer.OpDeployer, st deployer.OpaqueState, statePath string, idx int, l1ContractsVersion string) (*config.StagedChain, error) {
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

	l2Contracts, err := st.ReadL2ContractsLocator()
	if err != nil {
		return nil, fmt.Errorf("failed to read L2 contracts locator: %w", err)
	}
	if l2Contracts == "embedded" {
		l2Contracts = l1ContractsVersion
	}

	cfg := new(config.StagedChain)

	cfg.ChainID = uint64(common.HexToHash(chainId).Big().Int64())
	cfg.BatchInboxAddr = config.NewChecksummedAddress(dc.BatchInboxAddress)
	cfg.BlockTime = dc.L2BlockTime
	cfg.SeqWindowSize = dc.SequencerWindowSize
	cfg.MaxSequencerDrift = dc.MaxSequencerDrift
	cfg.DataAvailabilityType = "eth-da"
	cfg.DeploymentL1ContractsVersion = l1ContractsVersion
	cfg.DeploymentL2ContractsVersion = l2Contracts
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

	roles, err := GetRolesFromState(st, idx)
	if err != nil {
		return nil, fmt.Errorf("failed to read roles from state: %w", err)
	}

	// For TOML generation only include ProxyAdminOwner
	cfg.Roles = config.Roles{
		ProxyAdminOwner: roles.ProxyAdminOwner,
	}

	addresses, err := GetContractAddressesFromState(st, idx)
	if err != nil {
		return nil, fmt.Errorf("failed to read addresses from OpaqueState: %w", err)
	}

	cfg.Addresses = addresses

	// Check for depsets in the state.json
	interop, err := ExtractInteropDepSet(st)
	if err != nil {
		return nil, fmt.Errorf("failed to extract interop dep set from state: %w", err)
	}

	cfg.Interop = interop

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
	opcmAddress, err := st.ReadOpcmImpl()
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

func GetRolesFromState(st deployer.OpaqueState, idx int) (config.Roles, error) {
	roles := config.Roles{}

	systemConfigOwner, err := st.ReadSystemConfigOwner(idx)
	if err != nil {
		return roles, fmt.Errorf("failed to read system config owner: %w", err)
	}
	roles.SystemConfigOwner = config.NewChecksummedAddress(systemConfigOwner)

	proxyAdminOwner, err := st.ReadProxyAdminOwner(idx)
	if err != nil {
		return roles, fmt.Errorf("failed to read proxy admin owner: %w", err)
	}
	roles.ProxyAdminOwner = config.NewChecksummedAddress(proxyAdminOwner)

	guardian, err := st.ReadGuardian(idx)
	if err != nil {
		return roles, fmt.Errorf("failed to read guardian: %w", err)
	}
	roles.Guardian = config.NewChecksummedAddress(guardian)

	challenger, err := st.ReadChallenger(idx)
	if err != nil {
		return roles, fmt.Errorf("failed to read challenger: %w", err)
	}
	roles.Challenger = config.NewChecksummedAddress(challenger)

	proposer, err := st.ReadProposer(idx)
	if err != nil {
		return roles, fmt.Errorf("failed to read proposer: %w", err)
	}
	roles.Proposer = config.NewChecksummedAddress(proposer)

	unsafeBlockSigner, err := st.ReadUnsafeBlockSigner(idx)
	if err != nil {
		return roles, fmt.Errorf("failed to read unsafe block signer: %w", err)
	}
	roles.UnsafeBlockSigner = config.NewChecksummedAddress(unsafeBlockSigner)

	batchSubmitter, err := st.ReadBatchSubmitter(idx)
	if err != nil {
		return roles, fmt.Errorf("failed to read batch submitter: %w", err)
	}
	roles.BatchSubmitter = config.NewChecksummedAddress(batchSubmitter)

	return roles, nil
}

func GetContractAddressesFromState(st deployer.OpaqueState, idx int) (config.Addresses, error) {
	var addresses config.Addresses
	var err error

	l1StandardBridgeProxy, err := st.ReadL1StandardBridgeProxy(idx)
	if err != nil {
		return addresses, fmt.Errorf("failed to read L1StandardBridgeProxy: %w", err)
	}
	addresses.L1StandardBridgeProxy = config.NewChecksummedAddress(l1StandardBridgeProxy)

	optimismPortalProxy, err := st.ReadOptimismPortalProxy(idx)
	if err != nil {
		return addresses, fmt.Errorf("failed to read OptimismPortalProxy: %w", err)
	}
	addresses.OptimismPortalProxy = config.NewChecksummedAddress(optimismPortalProxy)

	systemConfigProxy, err := st.ReadSystemConfigProxy(idx)
	if err != nil {
		return addresses, fmt.Errorf("failed to read SystemConfigProxy: %w", err)
	}
	addresses.SystemConfigProxy = config.NewChecksummedAddress(systemConfigProxy)

	disputeGameFactoryProxy, err := st.ReadDisputeGameFactoryProxy(idx)
	if err != nil {
		return addresses, fmt.Errorf("failed to read DisputeGameFactoryProxy: %w", err)
	}
	addresses.DisputeGameFactoryProxy = config.NewChecksummedAddress(disputeGameFactoryProxy)

	return addresses, nil
}

// ExtractInteropDepSet reads the interop dependency set from state and converts it to config.Interop
func ExtractInteropDepSet(st deployer.OpaqueState) (*config.Interop, error) {
	interopDepSet, err := st.ReadInteropDepSet()
	if err != nil {
		return nil, fmt.Errorf("failed to read interop dep set: %w", err)
	}

	if interopDepSet == nil {
		return nil, nil
	}

	// Check if dependencies exists and is a map
	deps, ok := interopDepSet["dependencies"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("dependencies field is not a map or is missing")
	}

	// Return nil if dependencies map is empty
	if len(deps) == 0 {
		return nil, nil
	}

	interop := &config.Interop{
		Dependencies: make(map[string]config.StaticConfigDependency),
	}

	for key := range deps {
		interop.Dependencies[key] = config.StaticConfigDependency{}
	}

	return interop, nil
}
