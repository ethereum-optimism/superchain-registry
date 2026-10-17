package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

var (
	L1RPCURLsFlag = &cli.StringSliceFlag{
		Name:     "l1-rpc-urls",
		Usage:    "comma-separated list of L1 RPC URLs (only need multiple if checking multiple superchains)",
		EnvVars:  []string{"L1_RPC_URLS"},
		Required: true,
	}
	ChainIDFlag = &cli.Uint64SliceFlag{
		Name:  "chain-ids",
		Usage: "comma-separated list of l2 chainIds to check (optional, checks all chains if not provided)",
	}
	SuperchainsFlag = &cli.StringSliceFlag{
		Name:  "superchains",
		Usage: "comma-separated list of superchains to check (cannot provide both chain-ids and superchains flags, defaults to all superchains if not provided)",
	}
)

func main() {
	app := &cli.App{
		Name:  "check_onchain_configs",
		Usage: "read-only check that the [roles] and [addresses] in each chain's TOML config match onchain state",
		Flags: []cli.Flag{
			L1RPCURLsFlag,
			ChainIDFlag,
			SuperchainsFlag,
		},
		Action: CheckOnchainConfigsCLI,
	}
	if err := app.Run(os.Args); err != nil {
		output.WriteStderr("%v", err)
		os.Exit(1)
	}
}

func CheckOnchainConfigsCLI(cliCtx *cli.Context) error {
	l1RpcUrls := cliCtx.StringSlice("l1-rpc-urls")
	chainIds := cliCtx.Uint64Slice("chain-ids")
	var superchains []config.Superchain
	for _, sc := range cliCtx.StringSlice("superchains") {
		if sc != "" {
			superchains = append(superchains, sc)
		}
	}
	if len(chainIds) > 0 && len(superchains) > 0 {
		return fmt.Errorf("cannot provide both chain-ids and superchains flags")
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelError, false))
	wd, err := paths.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	diskCfgs, err := manage.CollectChainConfigs(paths.SuperchainConfigsDir(wd))
	if err != nil {
		return fmt.Errorf("error collecting chain configs: %w", err)
	}
	sort.Slice(diskCfgs, func(i, j int) bool {
		if diskCfgs[i].Superchain != diskCfgs[j].Superchain {
			return diskCfgs[i].Superchain < diskCfgs[j].Superchain
		}
		return diskCfgs[i].ShortName < diskCfgs[j].ShortName
	})

	onchainCfgs, err := manage.FetchChains(cliCtx.Context, lgr, wd, l1RpcUrls, chainIds, superchains)
	if err != nil {
		return fmt.Errorf("error fetching onchain configs: %w", err)
	}

	var numFailed int
	for _, cfg := range diskCfgs {
		onchainCfg, ok := onchainCfgs[cfg.Config.ChainID]
		if !ok {
			// Not selected by the chain-ids/superchains filters.
			continue
		}
		id := fmt.Sprintf("%s/%s (chain %d)", cfg.Superchain, cfg.ShortName, cfg.Config.ChainID)
		onchain := config.CreateAddressesWithRolesFromFetcher(onchainCfg.Addresses, onchainCfg.Roles)
		mismatches := manage.DiffChainAgainstOnchain(cfg.Config, onchain)
		var failed bool
		var numUnresolved int
		for _, m := range mismatches {
			if m.Unresolved() {
				output.WriteWarn("%s: %s (could not be resolved onchain, not treated as a failure)", id, m)
				numUnresolved++
				continue
			}
			output.WriteNotOK("%s: %s", id, m)
			failed = true
		}
		if failed {
			numFailed++
		} else if numUnresolved > 0 {
			output.WriteOK("%s has no mismatches (%d field(s) could not be resolved onchain)", id, numUnresolved)
		} else {
			output.WriteOK("%s matches onchain state", id)
		}
	}

	if numFailed > 0 {
		return fmt.Errorf("%d chain config(s) disagree with onchain state - update the [roles]/[addresses] in the listed superchain/configs/<superchain>/<chain>.toml files", numFailed)
	}
	return nil
}
