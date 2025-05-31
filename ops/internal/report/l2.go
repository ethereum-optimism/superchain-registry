package report

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum/go-ethereum/log"
)

// ScanL2 uses op-deployer to generate a standard genesis and diffs it against the provided genesis
func ScanL2(
	statePath string,
	l2ChainId uint64,
	l1RpcUrl string,
	deployerCacheDir string,
) (*L2Report, error) {
	st, err := deployer.ReadOpaqueStateFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read opaque state file: %w", err)
	}
	l1contractsrelease, err := st.ReadL1ContractsLocator()
	if err != nil {
		return nil, fmt.Errorf("failed to read L1 contracts release: %w", err)
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))
	opd, err := deployer.NewOpDeployer(lgr, l1contractsrelease, deployerCacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create op-deployer: %w", err)
	}

	originalGenesis, err := opd.InspectGenesis(statePath, strconv.FormatUint(l2ChainId, 10))
	if err != nil {
		return nil, fmt.Errorf("failed to inspect genesis: %w", err)
	}

	standardGenesis, err := opd.GenerateStandardGenesis(statePath, strconv.FormatUint(l2ChainId, 10), l1RpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to generate standard genesis: %w", err)
	}

	l2contractsrelease, err := st.ReadL2ContractsLocator()
	if err != nil {
		return nil, fmt.Errorf("failed to read L2 contracts release: %w", err)
	}

	if !strings.HasPrefix(l2contractsrelease, "tag://") {
		return nil, fmt.Errorf("L2 contracts release must have 'tag://' prefix, got: %s", l2contractsrelease)
	}
	tagValue := strings.TrimPrefix(l2contractsrelease, "tag://")

	var report L2Report
	report.Release = tagValue

	genesisDiffs := deployer.DiffOpaqueMaps("genesis", *originalGenesis, *standardGenesis)
	report.GenesisDiffs = genesisDiffs

	return &report, nil
}
