package main

import (
	"fmt"
	"os"

	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/urfave/cli/v2"
)

var FlagL1RPCURLs = &cli.StringSliceFlag{
	Name:     "l1-rpc-urls",
	Usage:    "Comma-separated list of L1 RPC URLs",
	Required: true,
}

func main() {
	app := &cli.App{
		Name:   "check-review-period-delays",
		Usage:  "checks that the review period delays of chains sharing an ETHLockbox are standard and consistent",
		Flags:  []cli.Flag{FlagL1RPCURLs},
		Action: CheckReviewPeriodDelaysCLI,
	}
	if err := app.Run(os.Args); err != nil {
		output.WriteStderr("%v", err)
		os.Exit(1)
	}
}

func CheckReviewPeriodDelaysCLI(cliCtx *cli.Context) error {
	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))
	wd, err := paths.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfgs, err := manage.CollectChainConfigs(paths.SuperchainConfigsDir(wd))
	if err != nil {
		return fmt.Errorf("failed to collect chain configs: %w", err)
	}

	superchainIds, err := paths.SuperchainIds(wd)
	if err != nil {
		return fmt.Errorf("failed to get superchain chain IDs: %w", err)
	}

	var addrs config.AddressesJSON
	if err := paths.ReadJSONFile(paths.AddressesFile(wd), &addrs); err != nil {
		return fmt.Errorf("failed to read addresses.json file: %w", err)
	}

	checker := manage.NewReviewPeriodChecker(lgr, cfgs, addrs, cliCtx.StringSlice(FlagL1RPCURLs.Name), superchainIds)
	if err := checker.Check(cliCtx.Context); err != nil {
		return fmt.Errorf("failed to validate review period delays: %w", err)
	}

	return nil
}
