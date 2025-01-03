package main

import (
	"fmt"
	"os"
	"path"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/inspect"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/fs"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/urfave/cli/v2"
)

var (
	FlagCheck = &cli.BoolFlag{
		Name:  "check",
		Usage: "Perform validation only.",
	}
	FlagPreserveInput = &cli.BoolFlag{
		Name:  "preserve-input",
		Usage: "Skip cleanup of staging directory.",
	}
)

func main() {
	app := &cli.App{
		Name:  "sync-staging",
		Usage: "Syncs a staging directory with the superchain registry.",
		Flags: []cli.Flag{
			FlagCheck,
			FlagPreserveInput,
		},
		Action: action,
	}
	if err := app.Run(os.Args); err != nil {
		output.WriteStderr("%v", err)
		os.Exit(1)
	}
}

func action(cliCtx *cli.Context) error {
	check := cliCtx.Bool(FlagCheck.Name)
	noCleanup := cliCtx.Bool(FlagPreserveInput.Name)
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	stagingDir := path.Join(wd, ".staging")

	intentExists, err := fs.FileExists(path.Join(stagingDir, "state.json"))
	if err != nil {
		return fmt.Errorf("failed to check for state file: %w", err)
	}
	if !intentExists {
		output.WriteOK("no staged intent file found, exiting")
		return nil
	}

	stagedChain, err := manage.NewStagedChain(stagingDir)
	if err != nil {
		return fmt.Errorf("failed to create staged chain: %w", err)
	}

	chainCfg, err := manage.InflateChainConfig(stagedChain)
	if err != nil {
		return fmt.Errorf("failed to inflate chain config: %w", err)
	}

	superchainPath := paths.SuperchainDir(wd, stagedChain.Meta.Superchain)
	chainCfgs, err := manage.CollectChainConfigs(superchainPath)
	if err != nil {
		return fmt.Errorf("failed to collect chain configs: %w", err)
	}

	if err := manage.ValidateUniqueness(chainCfg, stagedChain.Meta.ShortName, chainCfgs); err != nil {
		return fmt.Errorf("failed uniqueness check: %w", err)
	}
	output.WriteOK("internal uniqueness check passed")

	globalChainData, err := manage.FetchGlobalChainIDs()
	if err != nil {
		return fmt.Errorf("failed to fetch global chain IDs: %w", err)
	}
	if globalChainData.ChainIDs[chainCfg.ChainID] {
		return fmt.Errorf("chain ID %d is already in use", chainCfg.ChainID)
	}
	if globalChainData.ShortNames[stagedChain.Meta.ShortName] {
		return fmt.Errorf("short name %s is already in use", stagedChain.Meta.ShortName)
	}
	output.WriteOK("global uniqueness check passed")

	if check {
		output.WriteOK("validation successful")
		return nil
	}

	if err := manage.WriteChainConfig(wd, stagedChain.Meta, chainCfg); err != nil {
		return fmt.Errorf("failed to write chain config: %w", err)
	}

	output.WriteOK(
		"wrote chain config %s.toml to %s superchain",
		stagedChain.Meta.ShortName,
		stagedChain.Meta.Superchain,
	)

	genesis, _, err := inspect.GenesisAndRollup(stagedChain.State, stagedChain.State.AppliedIntent.Chains[0].ID)
	if err != nil {
		return fmt.Errorf("failed to get genesis: %w", err)
	}

	if err := manage.CompressGenesis(wd, stagedChain.Meta.Superchain, stagedChain.Meta.ShortName, genesis); err != nil {
		return fmt.Errorf("failed to compress genesis: %w", err)
	}

	output.WriteOK("wrote genesis files")

	if err := manage.GenAllCode(wd); err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	if !noCleanup {
		if err := stagedChain.Cleanup(); err != nil {
			return fmt.Errorf("failed to clean up staging directory: %w", err)
		}

		output.WriteOK("cleaned up staging directory")
	}

	return nil
}
