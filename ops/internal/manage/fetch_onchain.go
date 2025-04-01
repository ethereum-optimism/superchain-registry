package manage

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch"
	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch/script"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"golang.org/x/sync/errgroup"
)

// FetchSingleChain fetches configuration for a single chain by ID
func FetchSingleChain(lgr log.Logger, wd string, l1RpcUrls []string, chainIdStr string) (map[uint64]script.ChainConfig, error) {
	chainId, err := strconv.ParseUint(chainIdStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chainId: %w", err)
	}

	chainConfig, chainSuperchain, err := FindChainConfig(wd, chainId)
	if err != nil {
		return nil, err
	}

	l1RpcUrl, err := config.FindValidL1URL(lgr, l1RpcUrls, chainSuperchain)
	if err != nil {
		return nil, err
	}

	chainCfg, err := fetchChainInfo(context.Background(), lgr, l1RpcUrl, *chainConfig)
	if err != nil {
		return nil, err
	}

	lgr.Info("fetched chain config", "chainId", chainConfig.Config.ChainID)

	return map[uint64]script.ChainConfig{
		chainConfig.Config.ChainID: chainCfg,
	}, nil
}

// FetchAllSuperchains fetches data for all superchains that are compatible with the given l1RpcUrls
func FetchAllSuperchains(lgr log.Logger, wd string, l1RpcUrls []string) (map[uint64]script.ChainConfig, error) {
	superchains, err := paths.Superchains(wd)
	if err != nil {
		return nil, fmt.Errorf("error getting superchains: %w", err)
	}

	processedSuperchains := make(map[config.Superchain]bool)
	allChainConfigs := make(map[uint64]script.ChainConfig)

	for _, superchain := range superchains {
		if processedSuperchains[superchain] {
			continue
		}

		l1RpcUrl, err := config.FindValidL1URL(lgr, l1RpcUrls, superchain)
		if err != nil {
			lgr.Warn("skipping superchain - no valid L1 URL", "superchain", superchain, "error", err)
			continue
		}

		chains, err := fetchSuperchainConfigs(lgr, l1RpcUrl, wd, superchain)
		if err != nil {
			return nil, fmt.Errorf("error fetching configs for superchain %s: %w", superchain, err)
		}

		for chainID, config := range chains {
			allChainConfigs[chainID] = config
		}

		lgr.Info("fetched chain configs for superchain", "superchain", superchain, "chainCount", len(chains))
		processedSuperchains[superchain] = true
	}

	if len(processedSuperchains) == 0 {
		return nil, fmt.Errorf("no matching superchains found for any of the provided L1 RPC URLs")
	}

	lgr.Info("completed fetching data", "numSuperchains", len(processedSuperchains), "numChains", len(allChainConfigs))
	return allChainConfigs, nil
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

// fetchSuperchainConfigs fetches all onchain configs for a specific superchain
func fetchSuperchainConfigs(lgr log.Logger, l1RpcUrl, wd string, superchain config.Superchain) (map[uint64]script.ChainConfig, error) {
	cfgs, err := CollectChainConfigs(paths.SuperchainDir(wd, superchain))
	if err != nil {
		return nil, fmt.Errorf("error collecting chain configs: %w", err)
	}

	eg, ctx := errgroup.WithContext(context.Background())
	chainConfigs := make(map[uint64]script.ChainConfig)
	var mu sync.Mutex

	for _, c := range cfgs {
		cfg := c
		eg.Go(func() error {
			chainCfg, err := fetchChainInfo(ctx, lgr, l1RpcUrl, cfg)
			if err != nil {
				return err
			}

			mu.Lock()
			chainConfigs[cfg.Config.ChainID] = chainCfg
			mu.Unlock()

			lgr.Info("fetched chain config", "chainId", cfg.Config.ChainID)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return chainConfigs, nil
}
