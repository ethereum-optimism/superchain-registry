package report

import (
	"context"
	"time"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/artifacts"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

func ScanAll(
	ctx context.Context,
	rpcClient *rpc.Client,
	deploymentTx common.Hash,
	l1Release *artifacts.Locator,
	originalState *state.State,
) Report {
	var report Report
	var err error
	report.L1, err = ScanL1(
		ctx,
		rpcClient,
		deploymentTx,
		l1Release,
	)
	if err != nil {
		report.L1Err = err
	}

	report.L2, err = ScanL2(
		ctx,
		originalState,
	)
	if err != nil {
		report.L2Err = err
	}

	report.GeneratedAt = time.Now()
	return report
}
