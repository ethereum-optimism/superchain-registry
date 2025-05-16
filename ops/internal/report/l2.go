package report

import (
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum-optimism/optimism/op-chain-ops/addresses"
	"github.com/ethereum-optimism/optimism/op-chain-ops/foundry"
	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/broadcaster"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/opcm"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/env"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

func ScanL2(
	startBlock *types.Header,
	chainCfg *config.StagedChain,
	originalGenesis *core.Genesis,
	afacts foundry.StatDirFs,
) (*L2Report, error) {
	var report L2Report

	if !chainCfg.DeploymentL2ContractsVersion.IsTag() {
		return nil, errors.New("contracts version is not a tag")
	}

	report.Release = chainCfg.DeploymentL2ContractsVersion.Tag
	report.ProvidedGenesisHash = originalGenesis.ToBlock().Hash()

	standardGenesisHash, diffs, err := DiffL2Genesis(chainCfg, originalGenesis.Alloc, afacts, startBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to diff L2 genesis: %w", err)
	}

	report.StandardGenesisHash = standardGenesisHash
	report.AccountDiffs = diffs
	return &report, nil
}

func DiffL2Genesis(
	chainCfg *config.StagedChain,
	originalAllocs types.GenesisAlloc,
	artifacts foundry.StatDirFs,
	startBlock *types.Header,
) (common.Hash, []AccountDiff, error) {
	var standardHash common.Hash

	var l1ChainID uint64
	var standardRoles validation.RolesConfig
	if chainCfg.Superchain == config.MainnetSuperchain {
		standardRoles = validation.StandardConfigRolesMainnet
		l1ChainID = 1
	} else if chainCfg.Superchain == config.SepoliaSuperchain {
		standardRoles = validation.StandardConfigRolesSepolia
		l1ChainID = 11155111
	} else {
		return standardHash, nil, fmt.Errorf("unsupported superchain: %s", chainCfg.Superchain)
	}

	standardIntent := &state.Intent{
		ConfigType:         state.IntentTypeStandard,
		FundDevAccounts:    false,
		UseInterop:         false,
		L1ContractsLocator: chainCfg.DeploymentL1ContractsVersion,
		L2ContractsLocator: chainCfg.DeploymentL2ContractsVersion,
		L1ChainID:          l1ChainID,

		SuperchainRoles: &addresses.SuperchainRoles{
			SuperchainProxyAdminOwner: common.Address(standardRoles.L1ProxyAdminOwner),
			SuperchainGuardian:        common.Address(standardRoles.Guardian),
			ProtocolVersionsOwner:     common.Address(standardRoles.ProtocolVersionsOwner),
		},

		Chains: []*state.ChainIntent{
			{
				ID:                         common.BigToHash(new(big.Int).SetUint64(chainCfg.ChainID)),
				BaseFeeVaultRecipient:      common.Address(chainCfg.BaseFeeVaultRecipient),
				L1FeeVaultRecipient:        common.Address(chainCfg.L1FeeVaultRecipient),
				SequencerFeeVaultRecipient: common.Address(chainCfg.SequencerFeeVaultRecipient),
				Eip1559DenominatorCanyon:   chainCfg.Optimism.EIP1559DenominatorCanyon,
				Eip1559Denominator:         chainCfg.Optimism.EIP1559Denominator,
				Eip1559Elasticity:          chainCfg.Optimism.EIP1559Elasticity,
				Roles: state.ChainRoles{
					L1ProxyAdminOwner: common.Address(standardRoles.L1ProxyAdminOwner),
					L2ProxyAdminOwner: common.Address(standardRoles.L2ProxyAdminOwner),
					SystemConfigOwner: common.Address(*chainCfg.Roles.SystemConfigOwner),
					UnsafeBlockSigner: common.Address(*chainCfg.Roles.UnsafeBlockSigner),
					Batcher:           common.Address(*chainCfg.Roles.BatchSubmitter),
					Proposer:          common.Address(*chainCfg.Roles.Proposer),
					Challenger:        common.Address(*chainCfg.Roles.Challenger),
				},
			},
		},
	}

	// Hack until I find a better way of doing this
	if chainCfg.DeploymentL1ContractsVersion.Tag == string(validation.Semver160) {
		standardIntent.GlobalDeployOverrides = map[string]any{
			"l2GenesisHoloceneTimeOffset": nil,
		}
	}

	standardState := &state.State{
		// These values are not used in L2 genesis.
		SuperchainDeployment: new(addresses.SuperchainContracts),
	}

	opChainContracts := addresses.OpChainContracts{}
	opChainContracts.L1StandardBridgeProxy = common.Address(*chainCfg.Addresses.L1StandardBridgeProxy)
	opChainContracts.L1CrossDomainMessengerProxy = common.Address(*chainCfg.Addresses.L1CrossDomainMessengerProxy)
	opChainContracts.L1Erc721BridgeProxy = common.Address(*chainCfg.Addresses.L1ERC721BridgeProxy)
	opChainContracts.SystemConfigProxy = common.Address(*chainCfg.Addresses.SystemConfigProxy)
	opChainContracts.OptimismPortalProxy = common.Address(*chainCfg.Addresses.OptimismPortalProxy)

	standardChainState := &state.ChainState{
		ID: common.BigToHash(new(big.Int).SetUint64(chainCfg.ChainID)),
		StartBlock: &state.L1BlockRefJSON{
			Hash:       startBlock.Hash(),
			ParentHash: startBlock.ParentHash,
			Number:     hexutil.Uint64(startBlock.Number.Uint64()),
			Time:       hexutil.Uint64(startBlock.Time),
		},
		OpChainContracts: opChainContracts,
	}

	standardDeployConfig, err := state.CombineDeployConfig(standardIntent, standardIntent.Chains[0], standardState, standardChainState)
	if err != nil {
		return standardHash, nil, fmt.Errorf("failed to combine deploy config: %w", err)
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelWarn, false))
	deployer := common.Address{'D'}
	host, err := env.DefaultScriptHost(
		broadcaster.NoopBroadcaster(),
		lgr,
		deployer,
		artifacts,
	)
	if err != nil {
		return standardHash, nil, fmt.Errorf("failed to create script host: %w", err)
	}

	l2GenesisScript, err := opcm.NewL2GenesisScript(host)
	if err != nil {
		return standardHash, nil, fmt.Errorf("failed to call L2Genesis script: %w", err)
	}

	l1ChainId := new(big.Int).SetUint64(standardDeployConfig.L2InitializationConfig.L1ChainID)
	l2ChainId := new(big.Int).SetUint64(standardDeployConfig.L2InitializationConfig.L2ChainID)

	sequencerFeeVaultWithdrawalNetwork, baseFeeVaultWithdrawalNetwork, l1FeeVaultWithdrawalNetwork := l1ChainId, l1ChainId, l1ChainId
	if standardDeployConfig.L2InitializationConfig.SequencerFeeVaultWithdrawalNetwork == "local" {
		sequencerFeeVaultWithdrawalNetwork = l2ChainId
	}

	if standardDeployConfig.L2InitializationConfig.BaseFeeVaultWithdrawalNetwork == "local" {
		baseFeeVaultWithdrawalNetwork = l2ChainId
	}

	if standardDeployConfig.L2InitializationConfig.L1FeeVaultWithdrawalNetwork == "local" {
		l1FeeVaultWithdrawalNetwork = l2ChainId
	}

	if err := l2GenesisScript.Run(opcm.L2GenesisInput{
		L1ChainID:                                l1ChainId,
		L2ChainID:                                l2ChainId,
		L1CrossDomainMessengerProxy:              common.Address(*chainCfg.Addresses.L1CrossDomainMessengerProxy),
		L1StandardBridgeProxy:                    common.Address(*chainCfg.Addresses.L1StandardBridgeProxy),
		L1ERC721BridgeProxy:                      common.Address(*chainCfg.Addresses.L1ERC721BridgeProxy),
		OpChainProxyAdminOwner:                   standardDeployConfig.L2InitializationConfig.ProxyAdminOwner,
		SequencerFeeVaultRecipient:               standardDeployConfig.L2InitializationConfig.SequencerFeeVaultRecipient,
		SequencerFeeVaultMinimumWithdrawalAmount: standardDeployConfig.L2InitializationConfig.SequencerFeeVaultMinimumWithdrawalAmount.ToInt(),
		SequencerFeeVaultWithdrawalNetwork:       sequencerFeeVaultWithdrawalNetwork,
		BaseFeeVaultRecipient:                    standardDeployConfig.L2InitializationConfig.BaseFeeVaultRecipient,
		BaseFeeVaultMinimumWithdrawalAmount:      standardDeployConfig.L2InitializationConfig.BaseFeeVaultMinimumWithdrawalAmount.ToInt(),
		BaseFeeVaultWithdrawalNetwork:            baseFeeVaultWithdrawalNetwork,
		L1FeeVaultRecipient:                      standardDeployConfig.L2InitializationConfig.L1FeeVaultRecipient,
		L1FeeVaultMinimumWithdrawalAmount:        standardDeployConfig.L2InitializationConfig.L1FeeVaultMinimumWithdrawalAmount.ToInt(),
		L1FeeVaultWithdrawalNetwork:              l1FeeVaultWithdrawalNetwork,
		GovernanceTokenOwner:                     standardDeployConfig.L2InitializationConfig.GovernanceTokenOwner,
		Fork:                                     new(big.Int).SetUint64(uint64(standardDeployConfig.L2InitializationConfig.UpgradeScheduleDeployConfig.SolidityForkNumber(0))),
		UseInterop:                               standardDeployConfig.L2InitializationConfig.UseInterop,
		EnableGovernance:                         standardDeployConfig.L2InitializationConfig.EnableGovernance,
		FundDevAccounts:                          standardDeployConfig.L2InitializationConfig.FundDevAccounts,
	}); err != nil {
		return standardHash, nil, fmt.Errorf("failed to run L2Genesis script: %w", err)
	}

	host.Wipe(deployer)

	standardAllocs, err := host.StateDump()
	if err != nil {
		return standardHash, nil, fmt.Errorf("failed to dump state: %w", err)
	}

	standardGenesis, err := genesis.BuildL2Genesis(&standardDeployConfig, standardAllocs, &eth.BlockRef{
		Hash:       startBlock.Hash(),
		Number:     startBlock.Number.Uint64(),
		ParentHash: startBlock.ParentHash,
		Time:       startBlock.Time,
	})
	if err != nil {
		return standardHash, nil, fmt.Errorf("failed to build standard genesis: %w", err)
	}

	diffs := DiffAllocs(standardAllocs.Accounts, originalAllocs)
	return standardGenesis.ToBlock().Hash(), diffs, nil
}
