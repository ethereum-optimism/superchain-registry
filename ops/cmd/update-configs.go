package cmd

import (
	"path/filepath"
	"runtime"

	"github.com/ethereum-optimism/superchain-registry/ops/config"
	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/urfave/cli/v2"
)

var UpdateConfigsCmd = cli.Command{
	Name:    "update-configs",
	Aliases: []string{"u"},
	Usage:   "Update all config toml files after superchain.ChainConfig struct is updated",
	Action: func(ctx *cli.Context) error {
		for _, chain := range superchain.OPChains {
			_, thisFile, _, ok := runtime.Caller(0)
			if !ok {
				panic("Unable to get the current file path")
			}

			err := chain.CheckDataAvailability()
			if err != nil {
				panic(err)
			}

			superchainRepoPath := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
			targetDir := filepath.Join(superchainRepoPath, "superchain", "configs", chain.Superchain)
			targetFilePath := filepath.Join(targetDir, chain.Chain+".toml")
			err = config.WriteChainConfigTOML(*chain, targetFilePath)
			if err != nil {
				panic(err)
			}

		}
		return nil
	},
}
