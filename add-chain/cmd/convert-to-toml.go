package cmd

import (
	"fmt"
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
		err := config.WriteChainConfigTOML(*chain, targetFilePath)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Wrote file: %s\n", targetFilePath)
		return nil
	},
}
