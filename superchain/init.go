package superchain

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
)

func init() {
	superchainTargets, err := superchainFS.ReadDir("configs")
	if err != nil {
		panic(fmt.Errorf("failed to read superchain dir: %w", err))
	}
	// iterate over superchain-target entries
	for _, s := range superchainTargets {

		if !s.IsDir() {
			continue // ignore files, e.g. a readme
		}

		// Load superchain-target config
		superchainConfigData, err := superchainFS.ReadFile(path.Join("configs", s.Name(), "superchain.toml"))
		if err != nil {
			panic(fmt.Errorf("failed to read superchain config: %w", err))
		}
		var superchainEntry Superchain
		if err := unMarshalSuperchainConfig(superchainConfigData, &superchainEntry.Config); err != nil {
			panic(fmt.Errorf("failed to decode superchain config: %w", err))
		}
		superchainEntry.Superchain = s.Name()

		// iterate over the chains of this superchain-target
		chainEntries, err := superchainFS.ReadDir(path.Join("configs", s.Name()))
		if err != nil {
			panic(fmt.Errorf("failed to read superchain dir: %w", err))
		}
		for _, c := range chainEntries {
			if !isConfigFile(c) {
				continue
			}
			// load chain config
			chainConfigData, err := superchainFS.ReadFile(path.Join("configs", s.Name(), c.Name()))
			if err != nil {
				panic(fmt.Errorf("failed to read superchain config %s/%s: %w", s.Name(), c.Name(), err))
			}
			var chainConfig ChainConfig

			if err := toml.Unmarshal(chainConfigData, &chainConfig); err != nil {
				panic(fmt.Errorf("failed to decode chain config %s/%s: %w", s.Name(), c.Name(), err))
			}
			chainConfig.Chain = strings.TrimSuffix(c.Name(), ".toml")

			(&chainConfig).setNilHardforkTimestampsToDefaultOrZero(&superchainEntry.Config)

			MustBeValidSuperchainLevel(chainConfig)

			chainConfig.Superchain = s.Name()
			if other, ok := OPChains[chainConfig.ChainID]; ok {
				panic(fmt.Errorf("found chain config %q in superchain target %q with chain ID %d "+
					"conflicts with chain %q in superchain %q and chain ID %d",
					chainConfig.Name, chainConfig.Superchain, chainConfig.ChainID,
					other.Name, other.Superchain, other.ChainID))
			}
			superchainEntry.ChainIDs = append(superchainEntry.ChainIDs, chainConfig.ChainID)
			OPChains[chainConfig.ChainID] = &chainConfig
			Addresses[chainConfig.ChainID] = &chainConfig.Addresses
			GenesisSystemConfigs[chainConfig.ChainID] = &chainConfig.Genesis.SystemConfig
		}

		// Impute endpoints only if we're not in codegen mode.
		if os.Getenv("CODEGEN") == "" {
			runningInCI := os.Getenv("CI")
			switch superchainEntry.Superchain {
			case "mainnet":
				if runningInCI == "true" {
					superchainEntry.Config.L1.PublicRPC = "https://ci-mainnet-l1-archive.optimism.io"
				}
			case "sepolia", "sepolia-dev-0":
				if runningInCI == "true" {
					superchainEntry.Config.L1.PublicRPC = "https://ci-sepolia-l1-archive.optimism.io"
				}
			}
		}

		Superchains[superchainEntry.Superchain] = &superchainEntry
	}
}

func MustBeValidSuperchainLevel(chainConfig ChainConfig) {
	if chainConfig.SuperchainLevel != Frontier && chainConfig.SuperchainLevel != Standard {
		panic(fmt.Sprintf("invalid or unspecified superchain level %d", chainConfig.SuperchainLevel))
	}
}
