package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch"
	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch/script"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

var (
	SuperchainFlag = &cli.StringFlag{
		Name:     "superchain",
		Usage:    "Superchain to fetch.",
		Required: true,
	}
	L1RPCURLFlag = &cli.StringFlag{
		Name:     "l1-rpc-url",
		Usage:    "L1 RPC URL",
		Required: true,
	}
	OutputDirFlag = &cli.StringFlag{
		Name:     "output-dir",
		Usage:    "Output directory",
		Required: true,
	}
)

func main() {
	app := &cli.App{
		Name:  "fetch-onchain",
		Usage: "use op-fetcher to fetch onchain config data for all chains in a superchain",
		Flags: []cli.Flag{
			SuperchainFlag,
			L1RPCURLFlag,
			OutputDirFlag,
		},
		Action: FetchOnchainCLI,
	}
	if err := app.Run(os.Args); err != nil {
		output.WriteStderr("%v", err)
		os.Exit(1)
	}
}

func FetchOnchainCLI(cliCtx *cli.Context) error {
	l1RPCURL := cliCtx.String("l1-rpc-url")
	superchainFlag := cliCtx.String("superchain")
	outputDir := cliCtx.String("output-dir")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))

	superchain, err := config.ParseSuperchain(superchainFlag)
	if err != nil {
		return fmt.Errorf("error parsing superchain: %w", err)
	}

	if err := config.ValidateL1ChainID(l1RPCURL, superchain); err != nil {
		return fmt.Errorf("error validating L1 RPC URL: %w", err)
	}
	lgr.Info("l1-rpc-url is valid for superchain", "superchain", superchain)

	wd, err := paths.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("error finding repo root: %w", err)
	}

	cfgs, err := manage.CollectChainConfigs(paths.SuperchainDir(wd, config.Superchain(superchainFlag)))
	if err != nil {
		return fmt.Errorf("error collecting chain configs: %w", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(cfgs))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, cfg := range cfgs {
		wg.Add(1)
		go func(cfg *manage.DiskChainConfig) {
			defer wg.Done()
			if err := processChainConfig(ctx, lgr, cfg, outputDir, l1RPCURL); err != nil {
				errCh <- err
				cancel() // Signals to stop all running goroutines
			}
		}(&cfg)
	}

	wg.Wait()
	close(errCh)

	// If any errors occurred, return the first error
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	default:
	}

	lgr.Info("successfully fetched onchain data for all chains", "count", len(cfgs))
	return nil
}

// processChainConfig processes a single chain configuration and saves the result to a file
func processChainConfig(ctx context.Context, lgr log.Logger, cfg *manage.DiskChainConfig, outputDir, l1RPCURL string) error {
	fetcher, err := fetch.NewFetcher(
		lgr,
		l1RPCURL,
		common.HexToAddress(cfg.Config.Addresses.SystemConfigProxy.String()),
		common.HexToAddress(cfg.Config.Addresses.L1StandardBridgeProxy.String()),
	)
	if err != nil {
		return fmt.Errorf("error creating fetcher: %w", err)
	}

	result, err := fetcher.FetchChainInfo(ctx)
	if err != nil {
		return fmt.Errorf("error fetching chain info for chain %d: %w", cfg.Config.ChainID, err)
	}

	chainConfig := script.CreateChainConfig(result)
	filedata, err := json.MarshalIndent(chainConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal output for chain %d: %w", cfg.Config.ChainID, err)
	}

	filename := fmt.Sprintf("%d.json", cfg.Config.ChainID)
	outputFile := path.Join(outputDir, filename)
	err = os.WriteFile(outputFile, filedata, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write output to file for chain %d: %w", cfg.Config.ChainID, err)
	}

	lgr.Info("finished processing chain", "chain_id", cfg.Config.ChainID)
	return nil
}
