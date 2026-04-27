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
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3"
)

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

	deployedEvent, err := ParseDeployedEvent(receipt.Logs)
	if err != nil {
		return nil, fmt.Errorf("malformed Deployed event: %w", err)
	}

	// Fetch the transaction to get the To address
	tx, _, err := client.TransactionByHash(ctx, deploymentTx)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment transaction: %w", err)
	}

	// Validate that the transaction was sent to the expected OPCM address
	if tx.To() == nil {
		return nil, fmt.Errorf("deployment transaction has no To address (contract creation)")
	}
	if *tx.To() != opcmAddr {
		return nil, fmt.Errorf("unauthorized OPCM address: got %v, expected %v", tx.To(), opcmAddr)
	}
	output.WriteOK("deployment transaction was sent to the expected OPCM address: %v", opcmAddr)

	semversReport, err := ScanSemvers(ctx, rpcClient, deployedEvent.DeployOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to validate semvers: %w", err)
	}
	output.WriteOK("validated semvers")

	ownershipReport, err := ScanOwnership(ctx, rpcClient, deployedEvent.DeployOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ownership: %w", err)
	}
	output.WriteOK("validated ownership")

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

	// Set default gas paying token values
	report.IsGasPayingToken = false
	report.GasPayingToken = common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE")
	report.GasPayingTokenDecimals = 18
	report.GasPayingTokenName = "Ether"
	report.GasPayingTokenSymbol = "ETH"

	versionStr := strings.TrimPrefix(release, "op-contracts/")
	// Strip pre-release suffix (e.g., "-rc.2") to compare against base version
	if idx := strings.Index(versionStr, "-"); idx != -1 {
		versionStr = versionStr[:idx]
	}
	releaseSemver, err := semver.NewVersion(versionStr)
	if err != nil {
		return report, fmt.Errorf("failed to parse release: %w", err)
	}

	v180 := semver.MustParse("1.8.0")
	v200 := semver.MustParse("2.0.0")
	v500 := semver.MustParse("5.0.0")

	calls := []BatchCall{}

	// Always fetch these fields
	calls = append(
		calls,
		makeBatchCall(gasLimitABI, &report.GasLimit),
		makeBatchCall(scalarABI, &report.Scalar),
		makeBatchCall(overheadABI, &report.Overhead),
	)

	// Custom gas token functions removed in op-contracts/v2.0.0
	if releaseSemver.Compare(v200) < 0 {
		calls = append(
			calls,
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

	// For versions >= 1.8.0, fetch additional fields
	if releaseSemver.Compare(v180) >= 0 {
		calls = append(
			calls,
			makeBatchCall(baseFeeScalarABI, &report.BaseFeeScalar),
			makeBatchCall(blobBaseFeeScalarABI, &report.BlobBaseFeeScalar),
			makeBatchCall(eip1559DenominatorABI, &report.EIP1559Denominator),
			makeBatchCall(eip1559ElasticityABI, &report.EIP1559Elasticity),
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
