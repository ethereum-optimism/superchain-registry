package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

func ScanAll(
	ctx context.Context,
	l1RpcUrl string,
	rpcClient *rpc.Client,
	statePath string,
	chainCfg *config.StagedChain,
	deployerCacheDir string,
) Report {
	var report Report
	var err error

	l1ContractsRelease, err := GetContractsReleaseForOpcm(statePath)
	if err != nil {
		return Report{
			L1Err: err,
		}
	}

	var stateDeployOutput *DeployOPChainOutput
	if st, readErr := deployer.ReadOpaqueStateFile(statePath); readErr == nil {
		if deployOutput, outputErr := DeployOutputFromState(st, 0); outputErr == nil {
			stateDeployOutput = &deployOutput
		}
	}

	report.L1, err = ScanL1(
		ctx,
		rpcClient,
		*chainCfg.DeploymentTxHash,
		l1ContractsRelease,
		stateDeployOutput,
	)
	if err != nil {
		report.L1Err = err
	}

	report.L2, err = ScanL2(
		statePath,
		chainCfg.ChainID,
		l1RpcUrl,
		deployerCacheDir,
		l1ContractsRelease,
	)
	if err != nil {
		report.L2Err = err
	}

	report.GeneratedAt = time.Now()
	return report
}

func GetContractsReleaseForOpcm(statePath string) (string, error) {
	// Load state.json
	st, err := deployer.ReadOpaqueStateFile(statePath)
	if err != nil {
		return "", fmt.Errorf("failed to load state: %w", err)
	}

	// Read values from state
	opcmImplAddr, err := st.ReadOpcmImpl()
	if err != nil {
		return "", fmt.Errorf("failed to read OpcmImpl address from state: %w", err)
	}

	if opcmImplAddr == (common.Address{}) {
		return "", fmt.Errorf("OpcmImpl read from state is the zero address")
	}

	l1ChainID, err := st.ReadL1ChainID()
	if err != nil {
		return "", fmt.Errorf("failed to read L1 chain ID: %w", err)
	}

	// Select appropriate versions map based on L1 chain ID
	var versions validation.Versions
	switch l1ChainID {
	case 1: // Mainnet
		versions = validation.StandardVersionsMainnet
	case 11155111: // Sepolia
		versions = validation.StandardVersionsSepolia
	default:
		return "", fmt.Errorf("unsupported L1 chain ID: %d", l1ChainID)
	}

	return latestContractsReleaseForOpcm(opcmImplAddr, versions)
}

func latestContractsReleaseForOpcm(opcmAddr common.Address, versions validation.Versions) (string, error) {
	opcmAddrLower := strings.ToLower(opcmAddr.Hex())
	var latestTag string
	var latestVersion *semver.Version

	for versionTag, versionConfig := range versions {
		if versionConfig.OPContractsManager != nil &&
			versionConfig.OPContractsManager.Address != nil {
			if strings.ToLower(versionConfig.OPContractsManager.Address.String()) == opcmAddrLower {
				parsed, err := semver.NewVersion(strings.TrimPrefix(string(versionTag), "op-contracts/v"))
				if err != nil {
					return "", fmt.Errorf("failed to parse contracts version %s: %w", versionTag, err)
				}
				if latestVersion == nil || parsed.GreaterThan(latestVersion) {
					latestTag = string(versionTag)
					latestVersion = parsed
				}
			}
		}
	}

	if latestVersion == nil {
		return "", fmt.Errorf("OPCM address %s not found in standard versions", opcmAddr.Hex())
	}
	return latestTag, nil
}
