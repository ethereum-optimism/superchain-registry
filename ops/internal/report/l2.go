package report

import (
	"context"
	"fmt"
	"os"

	"github.com/ethereum-optimism/optimism/op-chain-ops/foundry"
	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/artifacts"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/broadcaster"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/inspect"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/opcm"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/env"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type ArtifactsProvider interface {
	Artifacts(ctx context.Context, loc *artifacts.Locator) (foundry.StatDirFs, error)
}

type artifactsHolder struct {
	fs  foundry.StatDirFs
	loc *artifacts.Locator
}

func (a *artifactsHolder) Artifacts(_ context.Context, loc *artifacts.Locator) (foundry.StatDirFs, error) {
	if !a.loc.Equal(loc) {
		return nil, fmt.Errorf("unexpected locator: %s", loc)
	}
	return a.fs, nil
}

func ScanL2(
	ctx context.Context,
	originalState *state.State,
) (*L2Report, error) {
	var report L2Report
	if originalState.AppliedIntent == nil {
		return nil, fmt.Errorf("no intent found in original state")
	}

	intent := originalState.AppliedIntent

	if !intent.L2ContractsLocator.IsTag() {
		return nil, fmt.Errorf("must use a tag for L2 contracts locator")
	}

	report.Release = intent.L2ContractsLocator.Tag

	originalGenesis, _, err := inspect.GenesisAndRollup(originalState, intent.Chains[0].ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate original genesis: %w", err)
	}

	report.ProvidedGenesisHash = originalGenesis.ToBlock().Hash()

	afacts, cleanup, err := artifacts.Download(
		ctx,
		intent.L2ContractsLocator,
		func(current, total int64) {
			output.WriteStderr("downloading L2 artifacts: %.2f/%.2f MB", float64(current)/1_000_000, float64(total)/1_000_000)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to download L2 artifacts: %w", err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			output.WriteStderr("failed to clean up L2 artifacts: %v", err)
		}
	}()
	artifactsProvider := &artifactsHolder{
		fs:  afacts,
		loc: intent.L2ContractsLocator,
	}

	standardGenesisHash, diffs, err := DiffL2Genesis(ctx, artifactsProvider, originalState)
	if err != nil {
		return nil, fmt.Errorf("failed to diff L2 genesis: %w", err)
	}

	report.StandardGenesisHash = standardGenesisHash
	report.AccountDiffs = diffs
	return &report, nil
}

func DiffL2Genesis(
	ctx context.Context,
	artifactsProvider ArtifactsProvider,
	originalState *state.State,
) (common.Hash, []AccountDiff, error) {
	var standardHash common.Hash
	originalIntent := originalState.AppliedIntent
	if originalIntent == nil {
		return standardHash, nil, fmt.Errorf("no intent found in original state")
	}

	if len(originalIntent.Chains) != 1 {
		return standardHash, nil, fmt.Errorf("expected exactly one chain in original intent, got %d", len(originalIntent.Chains))
	}
	if len(originalState.Chains) != 1 {
		return standardHash, nil, fmt.Errorf("expected exactly one chain in original state, got %d", len(originalState.Chains))
	}

	originalChainIntent := originalIntent.Chains[0]
	originalChainState := originalState.Chains[0]

	var roles validation.RolesConfig
	if originalIntent.L1ChainID == 1 {
		roles = validation.StandardConfigRolesMainnet
	} else if originalIntent.L1ChainID == 11155111 {
		roles = validation.StandardConfigRolesSepolia
	} else {
		return standardHash, nil, fmt.Errorf("unsupported L1 chain ID: %d", originalIntent.L1ChainID)
	}

	if ok := validation.IsValidContractSemver(originalIntent.L1ContractsLocator.Tag); !ok {
		return standardHash, nil, fmt.Errorf("invalid L1 contracts locator: %s", originalIntent.L1ContractsLocator)
	}
	if ok := validation.IsValidContractSemver(originalIntent.L2ContractsLocator.Tag); !ok {
		return standardHash, nil, fmt.Errorf("invalid L2 contracts locator: %s", originalIntent.L2ContractsLocator)
	}

	standardIntent := &state.Intent{
		DeploymentStrategy: state.DeploymentStrategyGenesis,
		ConfigType:         state.IntentConfigTypeStrict,
		FundDevAccounts:    false,
		UseInterop:         false,
		L1ContractsLocator: originalIntent.L1ContractsLocator,
		L2ContractsLocator: originalIntent.L2ContractsLocator,
		L1ChainID:          originalIntent.L1ChainID,

		SuperchainRoles: &state.SuperchainRoles{
			ProxyAdminOwner:       common.Address(roles.L1ProxyAdminOwner),
			Guardian:              common.Address(roles.Guardian),
			ProtocolVersionsOwner: common.Address(roles.ProtocolVersionsOwner),
		},

		Chains: []*state.ChainIntent{
			{
				ID:                         originalChainIntent.ID,
				BaseFeeVaultRecipient:      originalChainIntent.BaseFeeVaultRecipient,
				L1FeeVaultRecipient:        originalChainIntent.L1FeeVaultRecipient,
				SequencerFeeVaultRecipient: originalChainIntent.SequencerFeeVaultRecipient,
				Eip1559DenominatorCanyon:   originalChainIntent.Eip1559DenominatorCanyon,
				Eip1559Denominator:         originalChainIntent.Eip1559Denominator,
				Eip1559Elasticity:          originalChainIntent.Eip1559Elasticity,
				Roles: state.ChainRoles{
					L1ProxyAdminOwner: common.Address(roles.L1ProxyAdminOwner),
					L2ProxyAdminOwner: common.Address(roles.L2ProxyAdminOwner),
					SystemConfigOwner: originalChainIntent.Roles.SystemConfigOwner,
					UnsafeBlockSigner: originalChainIntent.Roles.UnsafeBlockSigner,
					Batcher:           originalChainIntent.Roles.Batcher,
					Proposer:          originalChainIntent.Roles.Proposer,
					Challenger:        originalChainIntent.Roles.Challenger,
				},
			},
		},
	}

	standardDeployConfig, err := state.CombineDeployConfig(standardIntent, standardIntent.Chains[0], originalState, originalChainState)
	if err != nil {
		return standardHash, nil, fmt.Errorf("failed to combine deploy config: %w", err)
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelWarn, false))

	artifactsL2, err := artifactsProvider.Artifacts(ctx, originalIntent.L2ContractsLocator)
	if err != nil {
		return standardHash, nil, fmt.Errorf("failed to download L2 artifacts: %w", err)
	}

	deployer := common.Address{'D'}
	host, err := env.DefaultScriptHost(
		broadcaster.NoopBroadcaster(),
		lgr,
		deployer,
		artifactsL2,
	)
	if err != nil {
		return standardHash, nil, fmt.Errorf("failed to create script host: %w", err)
	}

	if err := opcm.L2Genesis(host, &opcm.L2GenesisInput{
		L1Deployments: opcm.L1Deployments{
			L1CrossDomainMessengerProxy: originalChainState.L1CrossDomainMessengerProxyAddress,
			L1StandardBridgeProxy:       originalChainState.L1StandardBridgeProxyAddress,
			L1ERC721BridgeProxy:         originalChainState.L1ERC721BridgeProxyAddress,
		},
		L2Config: standardDeployConfig.L2InitializationConfig,
	}); err != nil {
		return standardHash, nil, fmt.Errorf("failed to call L2Genesis script: %w", err)
	}

	host.Wipe(deployer)

	standardAllocs, err := host.StateDump()
	if err != nil {
		return standardHash, nil, fmt.Errorf("failed to dump state: %w", err)
	}

	originalGenesis, _, err := inspect.GenesisAndRollup(originalState, originalChainIntent.ID)
	if err != nil {
		return standardHash, nil, fmt.Errorf("failed to generate original genesis: %w", err)
	}

	standardGenesis, err := genesis.BuildL2Genesis(&standardDeployConfig, standardAllocs, originalChainState.StartBlock)
	if err != nil {
		return standardHash, nil, fmt.Errorf("failed to build standard genesis: %w", err)
	}

	diffs := DiffAllocs(standardAllocs.Accounts, originalGenesis.Alloc)
	return standardGenesis.ToBlock().Hash(), diffs, nil
}
