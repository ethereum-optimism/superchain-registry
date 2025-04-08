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

	superchainIds, err := paths.SuperchainIds(wd)
	if err != nil {
		return nil, fmt.Errorf("error getting superchain chainIds: %w", err)
	}

	for superchain, chains := range chainsBySuperchain {
		lgr.Info("fetching superchain", "superchain", superchain, "numChains", len(chains))
		superchainId, ok := superchainIds[superchain]
		if !ok {
			return nil, fmt.Errorf("missing superchain chainId for superchain %s", superchain)
		}

		l1RpcUrl, err := config.FindValidL1URL(egCtx, lgr, l1RpcUrls, superchainId)
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
	result := make(map[config.Superchain][]DiskChainConfig)
	superchains, err := paths.Superchains(wd)
	if err != nil {
		return nil, fmt.Errorf("error getting superchains: %w", err)
	}

	// Create a map for quick chain ID lookup if we're filtering
	chainIdMap := make(map[uint64]bool)
	for _, id := range chainIds {
		chainIdMap[id] = true
	}

	foundChainIdMap := make(map[uint64]bool)

	// Process all superchains
	for _, superchain := range superchains {
		// Collect all chain configs from this superchain
		configs, err := CollectChainConfigs(paths.SuperchainDir(wd, superchain))
		if err != nil {
			return nil, fmt.Errorf("error collecting chain configs for superchain %s: %w", superchain, err)
		}

		// Filter configs if chainIds is specified
		if len(chainIds) > 0 {
			var filteredConfigs []DiskChainConfig
			for _, cfg := range configs {
				if chainIdMap[cfg.Config.ChainID] {
					filteredConfigs = append(filteredConfigs, cfg)
					foundChainIdMap[cfg.Config.ChainID] = true
				}
			}

			// Only add to map if we found matching chains for this superchain
			if len(filteredConfigs) > 0 {
				result[superchain] = filteredConfigs
			}
		} else {
			// If no chain IDs specified, include all chains
			result[superchain] = configs
		}
	}

	if len(chainIds) > 0 {
		// Ensure all requested chainIds were found
		for _, chainId := range chainIds {
			if !foundChainIdMap[chainId] {
				return nil, fmt.Errorf("chainId %d not found", chainId)
			}
		}
	}

	return result, nil
}

// fetchChainInfo handles the common logic for creating an op-fetcher instance using it to fetch chain info
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
