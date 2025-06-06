package main

import (
	"fmt"
	"os"

	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch/script"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

var (
	L1RPCURLsFlag = &cli.StringSliceFlag{
		Name:     "l1-rpc-urls",
		Usage:    "comma-separated list of L1 RPC URLs (only need multiple if fetching from multiple superchains)",
		EnvVars:  []string{"L1_RPC_URLS"},
		Required: true,
	}
	ChainIDFlag = &cli.Uint64SliceFlag{
		Name:  "chain-ids",
		Usage: "comma-separated list of l2 chainIds to update (optional, fetches all chains if not provided)",
	}
	SuperchainsFlag = &cli.StringSliceFlag{
		Name:  "superchains",
		Usage: "comma-separated list of superchains to update (cannot provide both chain-ids and superchains flags, default to all superchains if not provided)",
	}
)

func main() {
	app := &cli.App{
		Name:  "codegen",
		Usage: "uses op-fetcher to fetch onchain config data for chain(s) in the superchain-registry, then propagates the data to codegen files",
		Flags: []cli.Flag{
			L1RPCURLsFlag,
			ChainIDFlag,
			SuperchainsFlag,
		},
		Action: CodegenCLI,
	}
	if err := app.Run(os.Args); err != nil {
		output.WriteStderr("%v", err)
		os.Exit(1)
	}
}

func CodegenCLI(cliCtx *cli.Context) error {
	l1RpcUrls := cliCtx.StringSlice("l1-rpc-urls")
	chainIds := cliCtx.Uint64Slice("chain-ids")
	superchainsRaw := cliCtx.StringSlice("superchains")
	// Filter out empty strings from superchains
	var superchains []string
	for _, sc := range superchainsRaw {
		if sc != "" {
			superchains = append(superchains, sc)
		}
	}

	if len(chainIds) > 0 && len(superchains) > 0 {
		return fmt.Errorf("cannot provide both chain-ids and superchains flags")
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))
	wd, err := paths.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	var onchainCfgs map[uint64]script.ChainConfig
	ctx := cliCtx.Context
	onchainCfgs, err = manage.FetchChains(ctx, lgr, wd, l1RpcUrls, chainIds, superchains)
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
	return nil
}
