package manage

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch"
	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch/script"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"golang.org/x/sync/errgroup"
)

// FetchChains fetches chain configurations for specified chain IDs or all chains if none specified
func FetchChains(egCtx context.Context, lgr log.Logger, wd string, l1RpcUrls []string, chainIds []uint64) (map[uint64]script.ChainConfig, error) {
	chainsBySuperchain, err := collectChainsBySuperchain(wd, chainIds)
	if err != nil {
		return nil, err
	}

	allChainConfigs := make(map[uint64]script.ChainConfig)
	var mu sync.Mutex
	eg, egCtx := errgroup.WithContext(egCtx)

	for superchain, chains := range chainsBySuperchain {
		lgr.Info("fetching superchain", "superchain", superchain, "numChains", len(chains))
		l1RpcUrl, err := config.FindValidL1URL(egCtx, lgr, l1RpcUrls, superchain)
		if err != nil {
			return nil, fmt.Errorf("missing L1 RPC URL for superchain %s", superchain)
		}

		for _, cfg := range chains {
			// Capture variables for goroutine
			currentConfig := cfg
			currentRpcUrl := l1RpcUrl

			eg.Go(func() error {
				result, err := fetchChainInfo(egCtx, lgr, currentRpcUrl, currentConfig)
				if err != nil {
					return fmt.Errorf("failed to fetch chain info for chainId %d: %w", currentConfig.Config.ChainID, err)
				}

				mu.Lock()
				allChainConfigs[currentConfig.Config.ChainID] = result
				mu.Unlock()

				lgr.Info("fetched chain config", "chainId", currentConfig.Config.ChainID)
				return nil
			})
		}
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	lgr.Info("completed fetching", "totalChains", len(allChainConfigs))
	return allChainConfigs, nil
}

// collectChainsBySuperchain assembles a map of chains grouped by their superchain
// based on provided chain IDs or all available chains if no IDs are specified
func collectChainsBySuperchain(wd string, chainIds []uint64) (map[config.Superchain][]DiskChainConfig, error) {
	// Map to track chains by superchain
	chainsBySuperchain := make(map[config.Superchain][]DiskChainConfig)

	// Collect chains to process - either all chains or specific ones
	if len(chainIds) == 0 {
		// All chains mode
		superchains, err := paths.Superchains(wd)
		if err != nil {
			return nil, fmt.Errorf("error getting superchains: %w", err)
		}

		for _, superchain := range superchains {
			cfgs, err := CollectChainConfigs(paths.SuperchainDir(wd, superchain))
			if err != nil {
				return nil, fmt.Errorf("error collecting chain configs for superchain %s: %w", superchain, err)
			}
			chainsBySuperchain[superchain] = cfgs
		}
	} else {
		// Specific chains mode
		chainConfigTuples, err := FindChainConfigs(wd, chainIds)
		if err != nil {
			return nil, err
		}

		for _, tuple := range chainConfigTuples {
			chainsBySuperchain[tuple.Superchain] = append(chainsBySuperchain[tuple.Superchain], *tuple.Config)
		}
	}

	return chainsBySuperchain, nil
}

// fetchChainInfo handles the common logic for creating a fetcher and getting chain info
func fetchChainInfo(ctx context.Context, lgr log.Logger, l1RpcUrl string, cfg DiskChainConfig) (script.ChainConfig, error) {
	fetcher, err := fetch.NewFetcher(
		lgr,
		l1RpcUrl,
		common.HexToAddress(cfg.Config.Addresses.SystemConfigProxy.String()),
		common.HexToAddress(cfg.Config.Addresses.L1StandardBridgeProxy.String()),
	)
	if err != nil {
		return script.ChainConfig{}, fmt.Errorf("error creating fetcher for chain %d: %w", cfg.Config.ChainID, err)
	}

	result, err := fetcher.FetchChainInfo(ctx)
	if err != nil {
		return script.ChainConfig{}, fmt.Errorf("error fetching chain info for chain %d: %w", cfg.Config.ChainID, err)
	}

	return script.CreateChainConfig(result), nil
}
