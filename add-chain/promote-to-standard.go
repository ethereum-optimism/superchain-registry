package main

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/urfave/cli/v2"
)

var ChainIdFlag = &cli.Uint64Flag{
	Name:     "chain-id",
	Usage:    "ID of chain to promote",
	Required: true,
}

var PromoteToStandardCmd = cli.Command{
	Name:    "promote-to-standard",
	Flags:   []cli.Flag{ChainIdFlag},
	Aliases: []string{"p"},
	Usage:   "Promote a chain to standard.",
	Action: func(ctx *cli.Context) error {
		chainId := ChainIdFlag.Get(ctx)
		chain, ok := superchain.OPChains[chainId]
		if !ok {
			panic(fmt.Sprintf("No chain found with id %d", chainId))
		}

		chain.StandardChainCandidate = false
		chain.SuperchainLevel = superchain.Standard

		now := uint64(time.Now().Unix())
		chain.SuperchainTime = &now

		_, currentFilePath, _, ok := runtime.Caller(0)
		if !ok {
			panic("Unable to get the current file path")
		}

		superchainRepoPath := path.Join(currentFilePath, "../..")
		targetDir := filepath.Join(superchainRepoPath, "superchain", "configs", chain.Superchain)
		targetFilePath := filepath.Join(targetDir, chain.Name+".yaml")
		err := writeChainConfig(*chain, targetFilePath)
		if err != nil {
			panic(err)
		}

		fmt.Println("Promoted chain to standard: ", chainId)
		return nil
	},
}
