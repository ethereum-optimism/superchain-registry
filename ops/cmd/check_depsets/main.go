package main

import (
	"fmt"
	"os"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:   "check-depsets",
		Usage:  "checks that all interop depsets in the chain configs are valid",
		Action: CheckDepsetsCLI,
	}
	if err := app.Run(os.Args); err != nil {
		output.WriteStderr("%v", err)
		os.Exit(1)
	}
}

func CheckDepsetsCLI(cliCtx *cli.Context) error {
	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))
	wd, err := paths.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	superchainCfgsDir := paths.SuperchainConfigsDir(wd)
	cfgs, err := manage.CollectChainConfigs(superchainCfgsDir)
	if err != nil {
		return fmt.Errorf("failed to collect chain configs: %w", err)
	}

	var addrs config.AddressesJSON
	err = paths.ReadJSONFile(paths.AddressesFile(wd), &addrs)
	if err != nil {
		return fmt.Errorf("failed to read addresses.json file: %w", err)
	}

	checker, err := manage.NewDepsetChecker(lgr, cfgs, addrs)
	if err != nil {
		return fmt.Errorf("failed to create depset checker: %w", err)
	}
	if err := checker.Check(); err != nil {
		return fmt.Errorf("failed to validate depsets: %w", err)
	}

	return nil
}
