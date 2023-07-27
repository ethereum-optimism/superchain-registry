package superchain

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed configs
var superchainFS embed.FS

type SuperchainConfig struct {
	Name string `json:"name"`
}

func (cfg *SuperchainConfig) Check() error {
	// TODO
	return nil
}

type ChainConfig struct {
	Name string `json:"name"`
}

func (cfg *ChainConfig) Check() error {
	// TODO
	return nil
}

func init() {
	superchainTargets, err := superchainFS.ReadDir(".")
	if err != nil {
		panic(fmt.Errorf("failed to read superchain dir: %w", err))
	}
	// iterate over superchain-target entries
	for _, s := range superchainTargets {
		if !s.IsDir() {
			continue // ignore files, e.g. a readme
		}
		// TODO load superchain-target config
		//superchainConfigData, err := superchainFS.ReadFile("superchain.yaml")
		//if err != nil {
		//	panic(fmt.Errorf("failed to read superchain config: %w", err))
		//}

		// iterate over the chains of this superchain-target
		chainEntries, err := superchainFS.ReadDir(s.Name())
		if err != nil {
			panic(fmt.Errorf("failed to read superchain dir: %w", err))
		}
		for _, c := range chainEntries {
			if s.IsDir() || !strings.HasSuffix(c.Name(), ".yaml") {
				continue // ignore files. Chains must be a directory of configs.
			}
			if c.Name() != "superchain.yaml" {
				// process chain config
			}
			// TODO load chain config
		}
	}
}
