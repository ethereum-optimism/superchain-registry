package cmd

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ethereum-optimism/superchain-registry/add-chain/config"
	"github.com/ethereum-optimism/superchain-registry/add-chain/flags"
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

		chain.StandardChainCandidate = false
		chain.SuperchainLevel = superchain.Standard

		now := uint64(time.Now().Unix())
		chain.SuperchainTime = &now

		_, thisFile, _, ok := runtime.Caller(0)
		if !ok {
			panic("Unable to get the current file path")
		}

		superchainRepoPath := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
		targetDir := filepath.Join(superchainRepoPath, "superchain", "configs", chain.Superchain)
		targetFilePath := filepath.Join(targetDir, chain.Chain+".yaml")
		err := config.WriteChainConfig(*chain, targetFilePath)
		if err != nil {
			panic(err)
		}

		fmt.Println("Promoted chain to standard: ", chainId)
		return nil
	},
}
