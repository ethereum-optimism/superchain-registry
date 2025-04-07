package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch/script"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

var (
	L1RPCURLsFlag = &cli.StringFlag{
		Name:     "l1-rpc-urls",
		Usage:    "comma-separated list of L1 RPC URLs (only need multiple if fetching from multiple superchains)",
		EnvVars:  []string{"L1_RPC_URLS"},
		Required: true,
	}
	ChainIDFlag = &cli.StringFlag{
		Name:  "chain-ids",
		Usage: "comma-separated list of l2 chainIds to update (optional, fetches all chains if not provided)",
	}
)

func main() {
	app := &cli.App{
		Name:  "codegen",
		Usage: "uses op-fetcher to fetch onchain config data for chain(s) in the superchain-registry, then propagates the data to codegen files",
		Flags: []cli.Flag{
			L1RPCURLsFlag,
			ChainIDFlag,
		},
		Action: CodegenCLI,
	}
	if err := app.Run(os.Args); err != nil {
		output.WriteStderr("%v", err)
		os.Exit(1)
	}
}

func CodegenCLI(cliCtx *cli.Context) error {
	l1RpcUrls := strings.Split(cliCtx.String("l1-rpc-urls"), ",")
	chainIdStr := cliCtx.String("chain-ids")
	var chainIds []uint64
	if chainIdStr != "" {
		chainIdStrs := strings.Split(chainIdStr, ",")
		// Convert each string to uint64
		for _, idStr := range chainIdStrs {
			idStr = strings.TrimSpace(idStr)
			id, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid chain ID '%s': %w", idStr, err)
			}
			chainIds = append(chainIds, id)
		}
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))
	wd, err := paths.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	var onchainCfgs map[uint64]script.ChainConfig
	ctx := cliCtx.Context
	onchainCfgs, err = manage.FetchChains(ctx, lgr, wd, l1RpcUrls, chainIds)
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
