package manage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/inspect"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum/go-ethereum/common"
)

type StagedChain struct {
	State   *state.State
	Meta    *config.StagingMetadata
	cleanup func() error
}

func NewStagedChain(p string) (*StagedChain, error) {
	stateData, err := os.ReadFile(path.Join(p, "state.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read state: %w", err)
	}
	var st state.State
	if err := json.Unmarshal(stateData, &st); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	metaData, err := os.ReadFile(path.Join(p, "meta.toml"))
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}
	var meta config.StagingMetadata
	if err := toml.Unmarshal(metaData, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &StagedChain{
		State: &st,
		Meta:  &meta,
		cleanup: func() error {
			for _, f := range []string{"state.json", "meta.toml"} {
				if err := os.Remove(path.Join(p, f)); err != nil {
					return err
				}
			}
			return nil
		},
	}, nil
}

func (s *StagedChain) Cleanup() error {
	return s.cleanup()
}

func InflateChainConfig(sc *StagedChain) (*config.Chain, error) {
	chainIntent := sc.State.AppliedIntent.Chains[0]
	chainID := chainIntent.ID
	dc, err := inspect.DeployConfig(sc.State, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect deploy config: %w", err)
	}

	_, rollup, err := inspect.GenesisAndRollup(sc.State, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect genesis and rollup: %w", err)
	}

	cfg := new(config.Chain)
	cfg.Metadata = sc.Meta.Metadata
	cfg.ChainID = chainID.Big().Uint64()
	cfg.BatchInboxAddr = config.NewChecksummedAddress(dc.BatchInboxAddress)
	cfg.BlockTime = dc.L2BlockTime
	cfg.SeqWindowSize = dc.SequencerWindowSize
	cfg.MaxSequencerDrift = dc.MaxSequencerDrift

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
	}

	chainState := sc.State.Chains[0]
	cfg.Genesis = config.Genesis{
		L2Time: chainState.StartBlock.Time,
		L1: config.GenesisRef{
			Hash:   rollup.Genesis.L1.Hash,
			Number: rollup.Genesis.L1.Number,
		},
		L2: config.GenesisRef{
			Hash:   rollup.Genesis.L2.Hash,
			Number: rollup.Genesis.L2.Number,
		},
	}

	cfg.Roles = config.Roles{
		SystemConfigOwner: config.NewChecksummedAddress(chainIntent.Roles.SystemConfigOwner),
		ProxyAdminOwner:   config.NewChecksummedAddress(chainIntent.Roles.L1ProxyAdminOwner),
		Guardian:          config.NewChecksummedAddress(sc.State.AppliedIntent.SuperchainRoles.Guardian),
		Proposer:          config.NewChecksummedAddress(chainIntent.Roles.Proposer),
		UnsafeBlockSigner: config.NewChecksummedAddress(chainIntent.Roles.UnsafeBlockSigner),
		BatchSubmitter:    config.NewChecksummedAddress(chainIntent.Roles.Batcher),
	}

	cfg.Addresses = config.Addresses{
		AddressManager:                    config.NewChecksummedAddress(chainState.AddressManagerAddress),
		L1CrossDomainMessengerProxy:       config.NewChecksummedAddress(chainState.L1CrossDomainMessengerProxyAddress),
		L1ERC721BridgeProxy:               config.NewChecksummedAddress(chainState.L1ERC721BridgeProxyAddress),
		L1StandardBridgeProxy:             config.NewChecksummedAddress(chainState.L1StandardBridgeProxyAddress),
		OptimismMintableERC20FactoryProxy: config.NewChecksummedAddress(chainState.OptimismMintableERC20FactoryProxyAddress),
		OptimismPortalProxy:               config.NewChecksummedAddress(chainState.OptimismPortalProxyAddress),
		SystemConfigProxy:                 config.NewChecksummedAddress(chainState.SystemConfigProxyAddress),
		ProxyAdmin:                        config.NewChecksummedAddress(chainState.ProxyAdminAddress),
		SuperchainConfig:                  config.NewChecksummedAddress(sc.State.SuperchainDeployment.SuperchainConfigProxyAddress),
		AnchorStateRegistryProxy:          config.NewChecksummedAddress(chainState.AnchorStateRegistryProxyAddress),
		DelayedWETHProxy:                  config.NewChecksummedAddress(chainState.DelayedWETHPermissionedGameProxyAddress),
		DisputeGameFactoryProxy:           config.NewChecksummedAddress(chainState.DisputeGameFactoryProxyAddress),
		PermissionedDisputeGame:           config.NewChecksummedAddress(chainState.PermissionedDisputeGameAddress),
	}

	if dc.UseAltDA {
		cfg.Addresses.DAChallengeAddress = config.NewChecksummedAddress(dc.DAChallengeProxy)
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
		if !strings.HasPrefix(fieldName, "L2Genesis") ||
			!strings.HasSuffix(fieldName, "TimeOffset") {
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
