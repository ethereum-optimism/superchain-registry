package manage

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch"
	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch/script"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// FetchSingleChain fetches configuration for a single chain by ID
func FetchSingleChain(lgr log.Logger, l1RpcUrls []string, chainIdStr string) (map[uint64]script.ChainConfig, error) {
	chainId, err := strconv.ParseUint(chainIdStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chainId: %w", err)
	}

	repoRoot, err := paths.FindRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	chainConfig, chainSuperchain, err := FindChainConfig(repoRoot, chainId)
	if err != nil {
		return nil, err
	}

	l1RpcUrl, err := FindValidL1URL(lgr, l1RpcUrls, chainSuperchain)
	if err != nil {
		return nil, err
	}

	fetcher, err := fetch.NewFetcher(
		lgr,
		l1RpcUrl,
		common.HexToAddress(chainConfig.Config.Addresses.SystemConfigProxy.String()),
		common.HexToAddress(chainConfig.Config.Addresses.L1StandardBridgeProxy.String()),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating fetcher: %w", err)
	}

	result, err := fetcher.FetchChainInfo(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error fetching chain info for chain %d: %w", chainConfig.Config.ChainID, err)
	}

	chainCfg := script.CreateChainConfig(result)
	lgr.Info("fetched chain config", "chainId", chainConfig.Config.ChainID)

	return map[uint64]script.ChainConfig{
		chainConfig.Config.ChainID: chainCfg,
	}, nil
}

// FetchAllSuperchains fetches data for all superchains that are compatible with the given l1RpcUrls
func FetchAllSuperchains(lgr log.Logger, l1RpcUrls []string) (map[uint64]script.ChainConfig, error) {
	repoRoot, err := paths.FindRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("error finding repo root: %w", err)
	}

	superchains, err := paths.Superchains(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("error getting superchains: %w", err)
	}

	processedSuperchains := make(map[config.Superchain]bool)
	allChainConfigs := make(map[uint64]script.ChainConfig)

	for _, superchain := range superchains {
		if processedSuperchains[superchain] {
			continue
		}

		l1RpcUrl, err := FindValidL1URL(lgr, l1RpcUrls, superchain)
		if err != nil {
			lgr.Warn("skipping superchain - no valid L1 URL", "superchain", superchain, "error", err)
			continue
		}

		chains, err := fetchSuperchainConfigs(lgr, l1RpcUrl, repoRoot, superchain)
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

// FindChainConfig finds a chain configuration by chain ID
func FindChainConfig(repoRoot string, chainId uint64) (*DiskChainConfig, config.Superchain, error) {
	superchains, err := paths.Superchains(repoRoot)
	if err != nil {
		return nil, "", fmt.Errorf("error getting superchains: %w", err)
	}

	for _, superchain := range superchains {
		cfgs, err := CollectChainConfigs(paths.SuperchainDir(repoRoot, superchain))
		if err != nil {
			return nil, "", fmt.Errorf("error collecting chain configs: %w", err)
		}

		for _, cfg := range cfgs {
			if cfg.Config.ChainID == chainId {
				return &cfg, superchain, nil
			}
		}
	}

	return nil, "", fmt.Errorf("chain with id %d not found", chainId)
}

// FindValidL1URL finds a valid l1-rpc-url for a given superchain by finding matching l1 chainId
func FindValidL1URL(lgr log.Logger, urls []string, superchain config.Superchain) (string, error) {
	for i, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}

		if err := config.ValidateL1ChainID(url, superchain); err != nil {
			lgr.Warn("l1-rpc-url has mismatched l1 chainId", "urlIndex", i, "error", err)
			continue
		}

		lgr.Info("l1-rpc-url has matching l1 chainId", "urlIndex", i)
		return url, nil
	}
	return "", fmt.Errorf("no valid L1 RPC URL found for superchain %s", superchain)
}

// fetchSuperchainConfigs fetches all onchain configs for a specific superchain
func fetchSuperchainConfigs(lgr log.Logger, l1RpcUrl, repoRoot string, superchain config.Superchain) (map[uint64]script.ChainConfig, error) {
	cfgs, err := CollectChainConfigs(paths.SuperchainDir(repoRoot, superchain))
	if err != nil {
		return nil, fmt.Errorf("error collecting chain configs: %w", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	chainConfigs := make(map[uint64]script.ChainConfig)
	errChan := make(chan error, len(cfgs))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, c := range cfgs {
		wg.Add(1)
		go func(cfg DiskChainConfig) {
			defer wg.Done()
			fetcher, err := fetch.NewFetcher(
				lgr,
				l1RpcUrl,
				common.HexToAddress(cfg.Config.Addresses.SystemConfigProxy.String()),
				common.HexToAddress(cfg.Config.Addresses.L1StandardBridgeProxy.String()),
			)
			if err != nil {
				errChan <- fmt.Errorf("error creating fetcher for chain %d: %w", cfg.Config.ChainID, err)
				cancel()
				return
			}

			result, err := fetcher.FetchChainInfo(ctx)
			if err != nil {
				errChan <- fmt.Errorf("error fetching chain info for chain %d: %w", cfg.Config.ChainID, err)
				cancel()
				return
			}

			mu.Lock()
			chainCfg := script.CreateChainConfig(result)
			chainConfigs[cfg.Config.ChainID] = chainCfg
			mu.Unlock()

			lgr.Info("fetched chain config", "chainId", cfg.Config.ChainID)
		}(c)
	}

	wg.Wait()
	close(errChan)

	select {
	case err := <-errChan:
		if err != nil {
			return nil, err
		}
	default:
	}

	return chainConfigs, nil
}
