package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ethereum-optimism/superchain-registry/add-chain/config"
	"github.com/ethereum-optimism/superchain-registry/add-chain/flags"
	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/urfave/cli/v2"
)

var ConvertToTOMLCmd = cli.Command{
	Name:  "convert-to-toml",
	Flags: []cli.Flag{flags.ChainIdFlag},
	Usage: "Convert config yml to toml",
	Action: func(ctx *cli.Context) error {
		chainId := flags.ChainIdFlag.Get(ctx)
		chain, ok := superchain.OPChains[chainId]
		if !ok {
			panic(fmt.Sprintf("No chain found with id %d", chainId))
		}

		_, thisFile, _, ok := runtime.Caller(0)
		if !ok {
			panic("Unable to get the current file path")
		}

		superchainRepoPath := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
		targetDir := filepath.Join(superchainRepoPath, "superchain", "configs", chain.Superchain)
		targetFilePath := filepath.Join(targetDir, chain.Chain+".toml")

		var addressList superchain.AddressList
		addressesPath := filepath.Join(superchainRepoPath, "superchain", "extra", "addresses", chain.Superchain, chain.Chain+".json")
		rawData, err := os.ReadFile(addressesPath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		if err = json.Unmarshal(rawData, &addressList); err != nil {
			return fmt.Errorf("failed to unmarshal json: %w", err)
		}

		chain.Addresses = addressList

		var genesisSystemConfig superchain.SystemConfig
		genesisPath := filepath.Join(superchainRepoPath, "superchain", "extra", "genesis-system-configs", chain.Superchain, chain.Chain+".json")
		rawData, err = os.ReadFile(genesisPath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		if err = json.Unmarshal(rawData, &genesisSystemConfig); err != nil {
			return fmt.Errorf("failed to unmarshal json: %w", err)
		}

		chain.Addresses = addressList
		chain.Genesis.SystemConfig = genesisSystemConfig

		err = config.WriteChainConfigTOML(*chain, targetFilePath)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Wrote file: %s\n", targetFilePath)
		return nil
	},
}
