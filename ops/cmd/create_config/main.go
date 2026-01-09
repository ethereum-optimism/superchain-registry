package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/report"
	"github.com/urfave/cli/v2"
)

var (
	StateFilename = &cli.StringFlag{
		Name:      "state-filename",
		Usage:     "Filename of an op-deployer state file.",
		Required:  true,
		TakesFile: true,
	}
	Shortname = &cli.StringFlag{
		Name:     "shortname",
		Usage:    "Shortname of the chain.",
		Required: true,
		Value:    "newchain",
	}
	OpDeployerBinDir = &cli.StringFlag{
		Name:    "op-deployer-bin-dir",
		Usage:   "Path to the directory containing op-deployer binaries.",
		EnvVars: []string{"DEPLOYER_CACHE_DIR"},
		Value:   defaultBinDir(),
	}
	OpDeployerVersion = &cli.StringFlag{
		Name:    "op-deployer-version",
		Usage:   "Version of the op-deployer binary to use.",
		EnvVars: []string{"DEPLOYER_VERSION"},
		Value:   "",
	}
	L1ContractsVersion = &cli.StringFlag{
		Name:    "l1-contracts-version",
		Usage:   "Version tag of the L1 contracts (e.g., 'op-contracts/v1.6.0'). If not specified, will be auto-detected from state.json",
		EnvVars: []string{"L1_CONTRACTS_VERSION"},
		Value:   "",
	}
)

func main() {
	app := &cli.App{
		Name:  "create-config",
		Usage: "Turns a state file into a chain config in the stating directory.",
		Flags: []cli.Flag{
			StateFilename,
			Shortname,
			OpDeployerBinDir,
			OpDeployerVersion,
			L1ContractsVersion,
		},
		Action: action,
	}
	if err := app.Run(os.Args); err != nil {
		output.WriteStderr("%v", err)
		os.Exit(1)
	}
}

func action(cliCtx *cli.Context) error {
	wd, err := paths.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	statePath := cliCtx.String(StateFilename.Name)
	opDeployerBinDir := cliCtx.String(OpDeployerBinDir.Name)
	opDeployerVersion := cliCtx.String(OpDeployerVersion.Name)
	l1ContractsVersion := cliCtx.String(L1ContractsVersion.Name)

	if l1ContractsVersion == "" {
		var err error
		l1ContractsVersion, err = report.GetContractsReleaseForOpcm(statePath)
		if err != nil {
			return fmt.Errorf("failed to determine L1 contracts version: %w", err)
		}
	}

	output.WriteWarn("⚠️  Config generation behavior has changed: now generates only essential addresses by default.")
	output.WriteWarn("📄 All addresses are still available in addresses.json")

	err = manage.GenerateChainArtifacts(
		statePath,
		wd,
		cliCtx.String(Shortname.Name),
		nil, // name
		nil, // superchain
		0,   // idx
		opDeployerVersion,
		opDeployerBinDir,
		l1ContractsVersion,
	)
	if err != nil {
		return fmt.Errorf("failed to generate chain config: %w", err)
	}

	output.WriteOK("done")
	return nil
}

func defaultBinDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get home directory: %v", err))
	}

	return filepath.Join(homeDir, ".cache", "op-deployer")
}
