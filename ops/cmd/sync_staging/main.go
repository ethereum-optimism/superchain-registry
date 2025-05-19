package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/log"
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
	FlagL1RPCURLs = &cli.StringSliceFlag{
		Name:     "l1-rpc-urls",
		Usage:    "Comma-separated list of L1 RPC URLs",
		Required: true,
	}
)

func main() {
	app := &cli.App{
		Name:  "sync-staging",
		Usage: "Syncs a staging directory with the superchain registry.",
		Flags: []cli.Flag{
			FlagCheck,
			FlagPreserveInput,
			FlagL1RPCURLs,
		},
		Action: action,
	}
	if err := app.Run(os.Args); err != nil {
		output.WriteStderr("%v", err)
		os.Exit(1)
	}
}

func action(cliCtx *cli.Context) error {

	l1RpcUrls := cliCtx.StringSlice(FlagL1RPCURLs.Name)
	check := cliCtx.Bool(FlagCheck.Name)
	noCleanup := cliCtx.Bool(FlagPreserveInput.Name)
	wd, err := paths.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	stagingDir := paths.StagingDir(wd)

	stagedSuperchainDefinition, err := manage.StagedSuperchainDefinition(wd)
	stagedSuperchainDefinition.L1.PublicRPC = l1RpcUrls[0]

	if err == nil {
		output.WriteOK("superchain definition found, syncing...")
		err = manage.WriteSuperchainDefinition(
			path.Join(wd, "superchain", "configs", stagedSuperchainDefinition.Name, "superchain.toml"),
			stagedSuperchainDefinition)
		output.WriteOK("wrote superchain definition")
	} else if !errors.Is(err, manage.ErrNoStagedSuperchainDefinition) { // on this error we don't do anything
		return fmt.Errorf("failed to get staged superchain definition: %w", err)
	}

	if !noCleanup {
		superchainTomlPath := path.Join(stagingDir, "superchain.toml")
		if err := os.Remove(superchainTomlPath); err != nil {
			output.WriteNotOK("failed to remove %s: %v", superchainTomlPath, err)
		}
		output.WriteOK("cleaned superchain definition from staging directory")
	}

	stagedChainCfgs, err := manage.StagedChainConfigs(wd)
	if errors.Is(err, manage.ErrNoStagedConfig) {
		output.WriteOK("no staged chain config found, exiting")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get staged chain config: %w", err)
	}

	for _, chainCfg := range stagedChainCfgs {
		genesisFilename := path.Join(stagingDir, chainCfg.ShortName+".json.zst")
		genesis, err := manage.ReadGenesis(wd, genesisFilename)
		if err != nil {
			return fmt.Errorf("failed to read genesis: %w", err)
		}

		superchainPath := paths.SuperchainDir(wd, chainCfg.Superchain)
		chainCfgs, err := manage.CollectChainConfigs(superchainPath)
		if err != nil {
			return fmt.Errorf("failed to collect chain configs: %w", err)
		}

		if err := manage.ValidateUniqueness(chainCfg, chainCfgs); err != nil {
			return fmt.Errorf("failed uniqueness check: %w", err)
		}
		output.WriteOK("internal uniqueness check passed")

		if check {
			output.WriteOK("validation successful")
			continue
		}

		if err := manage.WriteChainConfig(wd, chainCfg); err != nil {
			return fmt.Errorf("failed to write chain config: %w", err)
		}

		output.WriteOK(
			"wrote chain config %s.toml to %s superchain",
			chainCfg.ShortName,
			chainCfg.Superchain,
		)

		if err := manage.WriteSuperchainGenesis(wd, chainCfg.Superchain, chainCfg.ShortName, genesis); err != nil {
			return fmt.Errorf("failed to write genesis: %w", err)
		}

		output.WriteOK("wrote genesis files")

		// Codegen
		lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))
		ctx := cliCtx.Context
		onchainCfgs, err := manage.FetchChains(ctx, lgr, wd, l1RpcUrls, []uint64{chainCfg.ChainID}, []config.Superchain{})
		if err != nil {
			return fmt.Errorf("error fetching onchain configs: %w", err)
		}
		syncer, err := manage.NewCodegenSyncer(lgr, wd, onchainCfgs)
		if err != nil {
			return fmt.Errorf("error creating codegen syncer: %w", err)
		}
		if err := syncer.SyncAll(); err != nil {
			return fmt.Errorf("error syncing codegen: %w", err)
		}

		if !noCleanup {
			cfgFilename := path.Join(stagingDir, chainCfg.ShortName+".toml")
			if err := os.Remove(cfgFilename); err != nil {
				output.WriteNotOK("failed to remove %s: %v", cfgFilename, err)
			}
			if err := os.Remove(genesisFilename); err != nil {
				output.WriteNotOK("failed to remove %s: %v", genesisFilename, err)
			}

			output.WriteOK("cleaned files from staging directory")
		}
	}

	return nil
}
