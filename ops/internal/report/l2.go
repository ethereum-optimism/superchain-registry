package report

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
)

// ScanL2 uses op-deployer to generate a standard genesis and diffs it against the provided genesis
func ScanL2(
	wd string,
	statePath string,
	l2ChainId uint64,
	l1RpcUrl string,
	deployerCacheDir string,
	l1ContractsVersion string,
) (*L2Report, error) {
	picker, err := deployer.WithReleaseBinary(deployerCacheDir, l1ContractsVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to autodetect binary: %w", err)
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))
	opd, err := deployer.NewOpDeployer(lgr, wd, picker)
	if err != nil {
		return nil, fmt.Errorf("failed to create op-deployer: %w", err)
	}

	originalGenesisJSON, err := opd.InspectGenesis(statePath, strconv.FormatUint(l2ChainId, 10))
	if err != nil {
		return nil, fmt.Errorf("failed to inspect genesis: %w", err)
	}
	var originalGenesis deployer.OpaqueMap
	if err := json.Unmarshal(originalGenesisJSON, &originalGenesis); err != nil {
		return nil, fmt.Errorf("failed to decode genesis: %w", err)
	}

	standardGenesis, err := opd.GenerateStandardGenesis(statePath, strconv.FormatUint(l2ChainId, 10), l1RpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to generate standard genesis: %w", err)
	}

	st, err := deployer.ReadOpaqueStateFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read opaque state file: %w", err)
	}

	l2contractsrelease, err := st.ReadL2ContractsLocator()
	if err != nil {
		return nil, fmt.Errorf("failed to read L2 contracts release: %w", err)
	}

	// Handle "embedded" or "tag://" cases
	var l2Release string
	if l2contractsrelease == "embedded" {
		// Use L1 version for L2 as well
		l2Release = l1ContractsVersion
	} else if strings.HasPrefix(l2contractsrelease, "tag://") {
		// Legacy format - strip prefix
		l2Release = strings.TrimPrefix(l2contractsrelease, "tag://")
	} else {
		// Assume it's already a clean version string
		l2Release = l2contractsrelease
	}

	var report L2Report
	report.Release = l2Release

	genesisDiffs := deployer.DiffOpaqueMaps("genesis", originalGenesis, *standardGenesis)
	report.GenesisDiffs = genesisDiffs

	return &report, nil
}
