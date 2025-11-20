package report

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/standard"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/gameargs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3"
)

type DeployedEvent struct {
	OutputVersion *big.Int
	L2ChainID     common.Hash
	Deployer      common.Address
	DeployOutput  DeployOPChainOutput
}

type DeployOPChainOutput struct {
	OpChainProxyAdmin                  common.Address
	AddressManager                     common.Address
	L1ERC721BridgeProxy                common.Address
	SystemConfigProxy                  common.Address
	OptimismMintableERC20FactoryProxy  common.Address
	L1StandardBridgeProxy              common.Address
	L1CrossDomainMessengerProxy        common.Address
	OptimismPortalProxy                common.Address
	DisputeGameFactoryProxy            common.Address
	AnchorStateRegistryProxy           common.Address
	AnchorStateRegistryImpl            common.Address
	FaultDisputeGame                   common.Address
	PermissionedDisputeGame            common.Address
	DelayedWETHPermissionedGameProxy   common.Address
	DelayedWETHPermissionlessGameProxy common.Address
}

func ParseDeployedEvent(log *types.Log) (*DeployedEvent, error) {
	outVersion := new(big.Int)
	var l2ChainID *big.Int
	var deployer common.Address
	var deployOutput []byte
	if err := deployedEventABI.DecodeArgs(log, &outVersion, &l2ChainID, &deployer, &deployOutput); err != nil {
		return nil, fmt.Errorf("failed to decode Deployed event: %w", err)
	}
	if outVersion.Cmp(common.Big0) != 0 {
		return nil, fmt.Errorf("unexpected output version: %v", outVersion)
	}

	deployedEv := &DeployedEvent{
		OutputVersion: outVersion,
		L2ChainID:     common.BigToHash(l2ChainID),
		Deployer:      deployer,
	}

	// Have to append the selector here since w3 only supports serializing
	// methods and events, not structs.
	deployOutputWithSel := make([]byte, 4+len(deployOutput))
	copy(deployOutputWithSel, deployOutputEvV0ABI.Selector[:])
	copy(deployOutputWithSel[4:], deployOutput)

	if err := deployOutputEvV0ABI.DecodeArgs(
		deployOutputWithSel,
		&deployedEv.DeployOutput.OpChainProxyAdmin,
		&deployedEv.DeployOutput.AddressManager,
		&deployedEv.DeployOutput.L1ERC721BridgeProxy,
		&deployedEv.DeployOutput.SystemConfigProxy,
		&deployedEv.DeployOutput.OptimismMintableERC20FactoryProxy,
		&deployedEv.DeployOutput.L1StandardBridgeProxy,
		&deployedEv.DeployOutput.L1CrossDomainMessengerProxy,
		&deployedEv.DeployOutput.OptimismPortalProxy,
		&deployedEv.DeployOutput.DisputeGameFactoryProxy,
		&deployedEv.DeployOutput.AnchorStateRegistryProxy,
		&deployedEv.DeployOutput.AnchorStateRegistryImpl,
		&deployedEv.DeployOutput.FaultDisputeGame,
		&deployedEv.DeployOutput.PermissionedDisputeGame,
		&deployedEv.DeployOutput.DelayedWETHPermissionedGameProxy,
		&deployedEv.DeployOutput.DelayedWETHPermissionlessGameProxy,
	); err != nil {
		return nil, fmt.Errorf("failed to decode deploy output: %w", err)
	}

	return deployedEv, nil
}

func ScanL1(
	ctx context.Context,
	rpcClient *rpc.Client,
	deploymentTx common.Hash,
	release string,
) (*L1Report, error) {
	client := ethclient.NewClient(rpcClient)

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	opcmAddr, err := standard.OPCMImplAddressFor(chainID.Uint64(), release)
	if err != nil {
		return nil, fmt.Errorf("failed to get OPCM address: %w", err)
	}

	receipt, err := client.TransactionReceipt(ctx, deploymentTx)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment receipt: %w", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("deployment tx failed: %v", receipt.Status)
	}

	var deploymentLog *types.Log
	for _, ev := range receipt.Logs {
		if ev.Topics[0] != deployedEventABI.Topic0 {
			continue
		}
		if deploymentLog != nil {
			return nil, fmt.Errorf("multiple Deployed events in receipt, this is unsupported")
		}
		deploymentLog = ev
	}
	if deploymentLog == nil {
		return nil, fmt.Errorf("no Deployed event in receipt")
	}
	if deploymentLog.Address != opcmAddr {
		return nil, fmt.Errorf("unauthorized address for Deployed event: %v", deploymentLog.Address)
	}

	deployedEvent, err := ParseDeployedEvent(deploymentLog)
	if err != nil {
		return nil, fmt.Errorf("malformed Deployed event: %w", err)
	}

	semversReport, err := ScanSemvers(ctx, rpcClient, deployedEvent.DeployOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to validate semvers: %w", err)
	}

	ownershipReport, err := ScanOwnership(ctx, rpcClient, deployedEvent.DeployOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ownership: %w", err)
	}

	permissionedGameReport, err := ScanFDG(
		ctx,
		rpcClient,
		1, // PERMISSIONED
		deployedEvent.DeployOutput.DisputeGameFactoryProxy,
		deployedEvent.DeployOutput.PermissionedDisputeGame,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to validate permissioned dispute game: %w", err)
	}

	systemConfigReport, err := ScanSystemConfig(ctx, rpcClient, release, deployedEvent.DeployOutput.SystemConfigProxy)
	if err != nil {
		return nil, fmt.Errorf("failed to validate system config: %w", err)
	}

	return &L1Report{
		Release:           release,
		DeploymentTxHash:  deploymentTx,
		DeploymentChainID: chainID.Uint64(),
		Semvers:           semversReport,
		Ownership:         ownershipReport,
		Proofs: L1ProofsReport{
			Permissioned: permissionedGameReport,
		},
		SystemConfig: systemConfigReport,
	}, nil
}

func ScanOwnership(
	ctx context.Context,
	rpc *rpc.Client,
	deployOutput DeployOPChainOutput,
) (L1OwnershipReport, error) {
	w3Client := w3.NewClient(rpc)

	var report L1OwnershipReport

	if err := CallBatch(
		ctx,
		w3Client,
		batchCallMethod(deployOutput.OptimismPortalProxy, guardianFnABI, &report.Guardian),
		batchCallMethod(deployOutput.PermissionedDisputeGame, challengerFnABI, &report.Challenger),
		batchCallMethod(deployOutput.OpChainProxyAdmin, ownerFnABI, &report.ProxyAdminOwner),
	); err != nil {
		return report, fmt.Errorf("failed to get ownership data: %w", err)
	}

	return report, nil
}

func ScanFDG(
	ctx context.Context,
	rpc *rpc.Client,
	gameType uint32,
	factoryAddr common.Address,
	gameAddr common.Address,
) (L1FDGReport, error) {
	w3Client := w3.NewClient(rpc)
	makeBatchCall := bindBatchCallTo(gameAddr)

	var maxGameDepth big.Int
	var splitDepth big.Int

	var report L1FDGReport
	if err := CallBatch(
		ctx,
		w3Client,
		makeBatchCall(gameTypeABI, &report.GameType),
		makeBatchCall(absolutePrestateFnABI, &report.AbsolutePrestate),
		makeBatchCall(maxGameDepthABI, &maxGameDepth),
		makeBatchCall(splitDepthABI, &splitDepth),
		makeBatchCall(maxClockDurationABI, &report.MaxClockDuration),
		makeBatchCall(clockExtensionABI, &report.ClockExtension),
	); err != nil {
		return report, fmt.Errorf("failed to get FDG data: %w", err)
	}

	if report.AbsolutePrestate == (common.Hash{}) {
		// Absolute prestate wasn't available from the implementation contract.
		// Try to fetch it from the game args instead.
		// Note that gameArgs isn't available in older DisputeGameFactory implementations so this isn't requested
		// as part of the above batch.
		var gameArgs []byte
		if err := CallBatch(
			ctx,
			w3Client,
			batchCallMethod(factoryAddr, gameArgsABI, &gameArgs, gameType),
		); err != nil {
			return report, fmt.Errorf("failed to get FDG game args: %w", err)
		}
		prestate, err := gameargs.ParseAbsoluteState(gameArgs)
		if err != nil {
			return report, fmt.Errorf("failed to parse FDG game args: %w", err)
		}
		report.AbsolutePrestate = prestate
		// When using game args, the game type is passed in by the DisputeGameFactory so always matches the game type
		// the implementation is set as.
		report.GameType = gameType
	}

	maxU64Big := new(big.Int).SetUint64(math.MaxUint64)
	if maxGameDepth.Cmp(maxU64Big) > 0 {
		return report, fmt.Errorf("unexpectedly large max game depth: %s", maxGameDepth.String())
	}
	if splitDepth.Cmp(maxU64Big) > 0 {
		return report, fmt.Errorf("unexpectedly large split depth: %s", splitDepth.String())
	}

	report.MaxGameDepth = maxGameDepth.Uint64()
	report.SplitDepth = splitDepth.Uint64()

	return report, nil
}

func ScanSystemConfig(
	ctx context.Context,
	rpc *rpc.Client,
	release string,
	addr common.Address,
) (L1SystemConfigReport, error) {
	w3Client := w3.NewClient(rpc)
	makeBatchCall := bindBatchCallTo(addr)
	var report L1SystemConfigReport

	versionStr := strings.TrimPrefix(release, "op-contracts/")
	releaseSemver, err := semver.NewVersion(versionStr)
	if err != nil {
		return report, fmt.Errorf("failed to parse release: %w", err)
	}

	v180 := semver.MustParse("1.8.0")
	v500 := semver.MustParse("5.0.0")

	calls := []BatchCall{}

	// Always fetch these base fields
	calls = append(
		calls,
		makeBatchCall(gasLimitABI, &report.GasLimit),
		makeBatchCall(scalarABI, &report.Scalar),
		makeBatchCall(overheadABI, &report.Overhead),
	)

	// For versions < 1.8.0, set default gas paying token values
	if releaseSemver.LessThan(v180) {
		report.IsGasPayingToken = false
		report.GasPayingToken = common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE")
		report.GasPayingTokenDecimals = 18
		report.GasPayingTokenName = "Ether"
		report.GasPayingTokenSymbol = "ETH"
	}

	// For versions >= 1.8.0, fetch additional fields
	if releaseSemver.Compare(v180) >= 0 {
		calls = append(
			calls,
			makeBatchCall(baseFeeScalarABI, &report.BaseFeeScalar),
			makeBatchCall(blobBaseFeeScalarABI, &report.BlobBaseFeeScalar),
			makeBatchCall(eip1559DenominatorABI, &report.EIP1559Denominator),
			makeBatchCall(eip1559ElasticityABI, &report.EIP1559Elasticity),
			makeBatchCall(isCustomGasTokenABI, &report.IsGasPayingToken),
			BatchCall{
				To: addr,
				Encoder: func() ([]byte, error) {
					return gasPayingTokenABI.EncodeArgs()
				},
				Decoder: func(rawOutput []byte) error {
					return gasPayingTokenABI.DecodeReturns(rawOutput, &report.GasPayingToken, &report.GasPayingTokenDecimals)
				},
			},
			makeBatchCall(gasPayingTokenNameABI, &report.GasPayingTokenName),
			makeBatchCall(gasPayingTokenSymbolABI, &report.GasPayingTokenSymbol),
		)
	}

	// For versions >= 5.0.0, fetch minBaseFee
	if releaseSemver.Compare(v500) >= 0 {
		calls = append(
			calls,
			makeBatchCall(minBaseFeeABI, &report.MinBaseFee),
		)
	}

	if err := CallBatch(
		ctx,
		w3Client,
		calls...,
	); err != nil {
		return report, fmt.Errorf("failed to get system config data: %w", err)
	}

	return report, nil
}

func ScanSemvers(
	ctx context.Context,
	rpc *rpc.Client,
	deployOutput DeployOPChainOutput,
) (L1SemversReport, error) {
	w3Client := w3.NewClient(rpc)
	makeBatchCall := bindBatchCallMethod(versionABI)

	var report L1SemversReport
	if err := CallBatch(
		ctx,
		w3Client,
		makeBatchCall(deployOutput.SystemConfigProxy, &report.SystemConfig),
		makeBatchCall(deployOutput.PermissionedDisputeGame, &report.PermissionedDisputeGame),
		makeBatchCall(deployOutput.OptimismPortalProxy, &report.OptimismPortal),
		makeBatchCall(deployOutput.AnchorStateRegistryProxy, &report.AnchorStateRegistry),
		makeBatchCall(deployOutput.DelayedWETHPermissionedGameProxy, &report.DelayedWETHPermissionedDisputeGame),
		makeBatchCall(deployOutput.DisputeGameFactoryProxy, &report.DisputeGameFactory),
		makeBatchCall(deployOutput.L1CrossDomainMessengerProxy, &report.L1CrossDomainMessenger),
		makeBatchCall(deployOutput.L1ERC721BridgeProxy, &report.L1ERC721Bridge),
		makeBatchCall(deployOutput.L1StandardBridgeProxy, &report.L1StandardBridge),
		makeBatchCall(deployOutput.OptimismMintableERC20FactoryProxy, &report.OptimismMintableERC20Factory),
	); err != nil {
		return report, fmt.Errorf("failed to get semvers data: %w", err)
	}

	return report, nil
}

func bindBatchCallTo(to common.Address) func(fn *w3.Func, out any, args ...any) BatchCall {
	return func(fn *w3.Func, out any, args ...any) BatchCall {
		return batchCallMethod(to, fn, out, args...)
	}
}

func bindBatchCallMethod(method *w3.Func) func(to common.Address, out any, args ...any) BatchCall {
	return func(to common.Address, out any, args ...any) BatchCall {
		return batchCallMethod(to, method, out, args...)
	}
}

func batchCallMethod(to common.Address, fn *w3.Func, out any, args ...any) BatchCall {
	return BatchCall{
		To: to,
		Encoder: func() ([]byte, error) {
			return fn.EncodeArgs(args...)
		},
		Decoder: func(rawOutput []byte) error {
			return fn.DecodeReturns(rawOutput, out)
		},
	}
}
