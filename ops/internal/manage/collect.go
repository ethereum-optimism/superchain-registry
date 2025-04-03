package manage

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/util"
)

const CollectorConcurrency = 8

type DiskChainConfig struct {
	ShortName string
	Filepath  string
	Config    *config.Chain
}

func CollectChainConfigs(p string) ([]DiskChainConfig, error) {
	var files []string

	err := filepath.Walk(p, func(fp string, info fs.FileInfo, err error) error {
		basePath := filepath.Base(fp)
		ext := filepath.Ext(basePath)
		if info.IsDir() || basePath == "superchain.toml" || ext != ".toml" {
			return nil
		}

		files = append(files, fp)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	filesCh := make(chan string)
	firstErr := new(util.OnceValue[error])
	var wg sync.WaitGroup
	var mtx sync.Mutex
	var out []DiskChainConfig

	worker := func() {
		defer wg.Done()
		for file := range filesCh {
			data, err := os.ReadFile(file)
			if err != nil {
				firstErr.Set(fmt.Errorf("failed to read file %s: %w", file, err))
				return
			}

			basename := filepath.Base(file)
			var chain config.Chain
			if err := toml.Unmarshal(data, &chain); err != nil {
				firstErr.Set(fmt.Errorf("failed to unmarshal toml %s: %w", basename, err))
				return
			}

			mtx.Lock()
			out = append(out, DiskChainConfig{
				ShortName: strings.TrimSuffix(basename, ".toml"),
				Filepath:  file,
				Config:    &chain,
			})
			mtx.Unlock()
		}
	}

	for i := 0; i < CollectorConcurrency; i++ {
		wg.Add(1)
		go worker()
	}

	for _, file := range files {
		filesCh <- file
	}
	close(filesCh)
	wg.Wait()

	if err := firstErr.V; err != nil {
		return nil, fmt.Errorf("error collecting configs: %w", err)
	}

	slices.SortFunc(out, func(a, b DiskChainConfig) int {
		return strings.Compare(a.Config.Name, b.Config.Name)
	})

	return out, nil
}

// FindChainConfig searches all superchains for given chain ID
func FindChainConfig(wd string, chainId uint64) (*DiskChainConfig, config.Superchain, error) {
	superchains, err := paths.Superchains(wd)
	if err != nil {
		return nil, "", fmt.Errorf("error getting superchains: %w", err)
	}

	for _, superchain := range superchains {
		cfgs, err := CollectChainConfigs(paths.SuperchainDir(wd, superchain))
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

type ChainConfigTuple struct {
	Config     *DiskChainConfig
	Superchain config.Superchain
}

// FindChainConfigs searches all superchains for the given chain IDs
// - returns an error if any chain ID is not found
func FindChainConfigs(wd string, chainIds []uint64) ([]ChainConfigTuple, error) {
	superchains, err := paths.Superchains(wd)
	if err != nil {
		return nil, fmt.Errorf("error getting superchains: %w", err)
	}

	// Create a map of chainIds for faster lookup
	remainingChainIds := make(map[uint64]bool)
	for _, id := range chainIds {
		remainingChainIds[id] = true
	}

	// Create a slice to store the results
	results := make([]ChainConfigTuple, 0, len(chainIds))

	// Search through all superchains for matching chain IDs
	for _, superchain := range superchains {
		// If we found all chains, we can exit early
		if len(remainingChainIds) == 0 {
			break
		}

		cfgs, err := CollectChainConfigs(paths.SuperchainDir(wd, superchain))
		if err != nil {
			return nil, fmt.Errorf("error collecting chain configs for superchain %s: %w", superchain, err)
		}

		for _, cfg := range cfgs {
			if remainingChainIds[cfg.Config.ChainID] {
				results = append(results, ChainConfigTuple{
					Config:     &cfg,
					Superchain: superchain,
				})

				delete(remainingChainIds, cfg.Config.ChainID)
			}
		}
	}

	if len(remainingChainIds) > 0 {
		return nil, fmt.Errorf("did not find the following chainIds: %v", remainingChainIds)
	}

	return results, nil
}
