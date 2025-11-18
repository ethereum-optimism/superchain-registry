package main

import (
	"fmt"
	"os"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

var ChainIDFlag = &cli.Uint64Flag{
	Name:     "chain-id",
	Usage:    "chain ID of the chain to remove",
	Required: true,
}

func main() {
	app := &cli.App{
		Name:  "remove-chain",
		Usage: "removes a chain config TOML and genesis file from the registry",
		Flags: []cli.Flag{
			ChainIDFlag,
		},
		Action: RemoveChainCLI,
	}
	if err := app.Run(os.Args); err != nil {
		output.WriteStderr("%v", err)
		os.Exit(1)
	}
}

func RemoveChainCLI(cliCtx *cli.Context) error {
	chainID := cliCtx.Uint64("chain-id")

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))
	wd, err := paths.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Collect all chain configs to find the one with matching chainID
	configs, err := manage.CollectChainConfigs(paths.SuperchainConfigsDir(wd))
	if err != nil {
		return fmt.Errorf("error collecting chain configs: %w", err)
	}

	var foundConfig *manage.DiskChainConfig
	for i := range configs {
		if configs[i].Config.ChainID == chainID {
			foundConfig = &configs[i]
			break
		}
	}

	if foundConfig == nil {
		return fmt.Errorf("chain with chain ID %d not found", chainID)
	}

	// Delete the chain config TOML file
	configPath := foundConfig.Filepath
	if err := os.Remove(configPath); err != nil {
		return fmt.Errorf("error removing chain config file %s: %w", configPath, err)
	}
	output.WriteOK("removed chain config: %s", configPath)

	// Delete the genesis file
	genesisPath := paths.GenesisFile(wd, foundConfig.Superchain, foundConfig.ShortName)
	if err := os.Remove(genesisPath); err != nil {
		// If genesis file doesn't exist, that's okay - just log a warning
		if !os.IsNotExist(err) {
			return fmt.Errorf("error removing genesis file %s: %w", genesisPath, err)
		}
		lgr.Warn("genesis file does not exist, skipping", "path", genesisPath)
	} else {
		output.WriteOK("removed genesis file: %s", genesisPath)
	}

	return nil
}
