package report

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

// ScanL2 uses op-deployer to generate a standard genesis and diffs it against the provided genesis
func ScanL2_2(
	startBlock *types.Header,
	statePath string,
	chainCfg *config.StagedChain,
	originalGenesis *core.Genesis,
) (*L2Report, error) {
	// Steps
	// 1. create a new op-deployer instance, opd
	// 2. use opd to create a new, standard genesis
	//    - standardGenesis := opd.GenerateStandardGenesis(statePath,workdir)
	// 3. diff the two geneses

	st, err := deployer.ReadOpaqueMappingFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read opaque mapping file: %w", err)
	}
	l1contractsrelease, err := st.ReadL1ContractsLocator()
	if err != nil {
		return nil, fmt.Errorf("failed to read L1 contracts release: %w", err)
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))
	opd, err := deployer.NewOpDeployer(lgr, l1contractsrelease)
	if err != nil {
		return nil, fmt.Errorf("failed to create op-deployer: %w", err)
	}

	standardGenesis, err := opd.GenerateStandardGenesis(statePath, strconv.FormatUint(chainCfg.ChainID, 10))
	if err != nil {
		return nil, fmt.Errorf("failed to generate standard genesis: %w", err)
	}

	if !chainCfg.DeploymentL2ContractsVersion.IsTag() {
		return nil, errors.New("contracts version is not a tag")
	}

	var report L2Report
	report.Release = chainCfg.DeploymentL2ContractsVersion.Tag
	report.ProvidedGenesisHash = originalGenesis.ToBlock().Hash()
	report.StandardGenesisHash = standardGenesis.ToBlock().Hash()
	report.AccountDiffs = DiffAllocs(standardGenesis.Alloc, originalGenesis.Alloc)

	return &report, nil
}
