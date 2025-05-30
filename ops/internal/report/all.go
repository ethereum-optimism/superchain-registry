package report

import (
	"context"
	"time"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/rpc"
)

func ScanAll(
	ctx context.Context,
	l1RpcUrl string,
	rpcClient *rpc.Client,
	statePath string,
	chainCfg *config.StagedChain,
	originalGenesis *core.Genesis,
) Report {
	var report Report
	var err error
	report.L1, err = ScanL1(
		ctx,
		rpcClient,
		*chainCfg.DeploymentTxHash,
		chainCfg.DeploymentL1ContractsVersion,
	)
	if err != nil {
		report.L1Err = err
	}
	report.L2, err = ScanL2(
		statePath,
		chainCfg.ChainID,
		l1RpcUrl,
	)
	if err != nil {
		report.L2Err = err
	}

	report.GeneratedAt = time.Now()
	return report
}
