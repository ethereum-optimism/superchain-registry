package cmd

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/ethereum-optimism/superchain-registry/ops/config"
	"github.com/ethereum-optimism/superchain-registry/ops/flags"
	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/urfave/cli/v2"
)

var PromoteToStandardCmd = cli.Command{
	Name:    "promote-to-standard",
	Flags:   []cli.Flag{flags.ChainIdFlag},
	Aliases: []string{"p"},
	Usage:   "Promote a chain to standard.",
	Action: func(ctx *cli.Context) error {
		chainId := flags.ChainIdFlag.Get(ctx)
		chain, ok := superchain.OPChains[chainId]
		if !ok {
			panic(fmt.Sprintf("No chain found with id %d", chainId))
		}

		copy, err := chain.PromoteToStandard()
		if err != nil {
			panic(err)
		}
		chain = copy

		_, thisFile, _, ok := runtime.Caller(0)
		if !ok {
			panic("Unable to get the current file path")
		}

		superchainRepoPath := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
		targetDir := filepath.Join(superchainRepoPath, "superchain", "configs", chain.Superchain)
		targetFilePath := filepath.Join(targetDir, chain.Chain+".toml")
		err = config.WriteChainConfigTOML(*chain, targetFilePath)
		if err != nil {
			panic(err)
		}

		fmt.Println("Promoted chain to standard: ", chainId)
		return nil
	},
}
