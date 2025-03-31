package manage

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch"
	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch/script"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

type OnchainFetcher struct {
	lgr      log.Logger
	l1RPCURL string
}

func NewOnchainFetcher(lgr log.Logger, l1RPCURL string) *OnchainFetcher {
	return &OnchainFetcher{
		lgr:      lgr,
		l1RPCURL: l1RPCURL,
	}
}

// FetchChainConfig fetches configuration for a single chain
func (f *OnchainFetcher) FetchChainConfig(ctx context.Context, cfg *DiskChainConfig) (script.ChainConfig, error) {
	fetcher, err := fetch.NewFetcher(
		f.lgr,
		f.l1RPCURL,
		common.HexToAddress(cfg.Config.Addresses.SystemConfigProxy.String()),
		common.HexToAddress(cfg.Config.Addresses.L1StandardBridgeProxy.String()),
	)
	if err != nil {
		return script.ChainConfig{}, fmt.Errorf("error creating fetcher: %w", err)
	}

	result, err := fetcher.FetchChainInfo(ctx)
	if err != nil {
		return script.ChainConfig{}, fmt.Errorf("error fetching chain info for chain %d: %w", cfg.Config.ChainID, err)
	}

	chainConfig := script.CreateChainConfig(result)
	f.lgr.Info("fetched chain config", "chainId", cfg.Config.ChainID)
	return chainConfig, nil
}

// FetchAllChainConfigs fetches configurations for all chains in a superchain
func (f *OnchainFetcher) FetchAllChainConfigs(ctx context.Context, superchain string) (map[uint64]script.ChainConfig, error) {
	wd, err := paths.FindRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("error finding repo root: %w", err)
	}

	cfgs, err := CollectChainConfigs(paths.SuperchainDir(wd, config.Superchain(superchain)))
	if err != nil {
		return nil, fmt.Errorf("error collecting chain configs: %w", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex // Protects chainConfigs from concurrent writes
	errCh := make(chan error, len(cfgs))
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	chainConfigs := make(map[uint64]script.ChainConfig)
	for _, cfg := range cfgs {
		wg.Add(1)
		go func(cfg *DiskChainConfig) {
			defer wg.Done()

			chainConfig, err := f.FetchChainConfig(ctxWithCancel, cfg)
			if err != nil {
				errCh <- fmt.Errorf("error fetching chain %d: %w", cfg.Config.ChainID, err)
				cancel() // Signals to stop all running goroutines
				return
			}

			mu.Lock()
			chainConfigs[cfg.Config.ChainID] = chainConfig
			mu.Unlock()
		}(&cfg)
	}

	wg.Wait()
	close(errCh)

	// If any errors occurred, return the first error
	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	default:
	}

	return chainConfigs, nil
}

func FetchSingleChain(lgr log.Logger, l1RpcUrls []string, chainIdStr string) (map[uint64]script.ChainConfig, error) {
	chainId, err := strconv.ParseUint(chainIdStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain-id: %w", err)
	}

	repoRoot, err := paths.FindRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	superchains, err := paths.Superchains(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("error getting superchains: %w", err)
	}

	var chainConfig *DiskChainConfig
	var chainSuperchain config.Superchain
	// Loop through all superchains until we find the L2 with matching chainId
superchainLoop:
	for _, superchain := range superchains {
		cfgs, err := CollectChainConfigs(paths.SuperchainDir(repoRoot, superchain))
		if err != nil {
			return nil, fmt.Errorf("error collecting chain configs: %w", err)
		}

		for _, cfg := range cfgs {
			if cfg.Config.ChainID == chainId {
				chainConfig = &cfg
				chainSuperchain = superchain
				break superchainLoop
			}
		}
	}

	if chainConfig == nil {
		return nil, fmt.Errorf("chain with id %d not found", chainId)
	}

	var matchingL1RpcUrl string
	for i, l1RpcUrl := range l1RpcUrls {
		if err := config.ValidateL1ChainID(l1RpcUrl, chainSuperchain); err != nil {
			lgr.Warn("l1-rpc-url has unexpected l1 chainId", "urlIndex", i)
			continue
		}
		matchingL1RpcUrl = l1RpcUrl
		lgr.Info("l1-rpc-url has matching l1 chainId", "urlIndex", i)
		break
	}
	if matchingL1RpcUrl == "" {
		return nil, fmt.Errorf("no valid L1 RPC URL found for superchain %s", chainSuperchain)
	}

	onchainFetcher := NewOnchainFetcher(lgr, matchingL1RpcUrl)
	onchainCfg, err := onchainFetcher.FetchChainConfig(context.Background(), chainConfig)
	if err != nil {
		return nil, fmt.Errorf("error fetching onchain configs: %w", err)
	}

	onchainCfgs := map[uint64]script.ChainConfig{
		chainConfig.Config.ChainID: onchainCfg,
	}
	return onchainCfgs, nil
}

// FetchAllSuperchains fetches data for all superchains using the provided L1 RPC URLs
// and returns a combined map of all L2 chain configurations
func FetchAllSuperchains(lgr log.Logger, l1RpcUrls []string) (map[uint64]script.ChainConfig, error) {
	processedSuperchains := make(map[config.Superchain]bool)
	allChainConfigs := make(map[uint64]script.ChainConfig)

	for i, l1RpcUrl := range l1RpcUrls {
		l1RpcUrl = strings.TrimSpace(l1RpcUrl)
		if l1RpcUrl == "" {
			continue
		}

		l1ChainID, err := getL1ChainID(l1RpcUrl)
		if err != nil {
			lgr.Warn("failed to get chain ID from L1 RPC URL", "urlIndex", i)
			continue
		}
		lgr.Info("detected L1 chain ID", "urlIndex", i, "chain_id", l1ChainID)

		for superchain, expectedChainID := range config.SuperchainChainIds {
			if processedSuperchains[superchain] {
				continue
			}

			if expectedChainID == l1ChainID {
				lgr.Info("found matching superchain", "superchain", superchain, "l1_chain_id", l1ChainID)
				onchainFetcher := NewOnchainFetcher(lgr, l1RpcUrl)
				superchainConfigs, err := onchainFetcher.FetchAllChainConfigs(context.Background(), string(superchain))
				if err != nil {
					return nil, fmt.Errorf("error fetching onchain configs for superchain %s: %w", superchain, err)
				}

				// Add all configs from this superchain to our combined result map
				for chainID, config := range superchainConfigs {
					allChainConfigs[chainID] = config
				}

				lgr.Info("successfully fetched onchain data for superchain", "superchain", superchain, "chain_count", len(superchainConfigs))
				processedSuperchains[superchain] = true
			}
		}
	}

	// Check if we processed any superchains
	if len(processedSuperchains) == 0 {
		return nil, fmt.Errorf("no matching superchains found for any of the provided L1 RPC URLs")
	}

	lgr.Info("completed fetching data for all matching superchains", "numSuperchains", len(processedSuperchains), "numChains", len(allChainConfigs))
	return allChainConfigs, nil
}

func getL1ChainID(rpcURL string) (uint64, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to L1 RPC: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return chainID.Uint64(), nil
}
