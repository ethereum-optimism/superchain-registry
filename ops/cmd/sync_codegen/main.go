package main

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "sync-codegen"
	app.Usage = "Synchronize codegen files with data fetched from onchain"
	app.Description = "Updates addresses.json and chainList.json files with data from individual chain JSON files"
	app.Action = syncCodegenFiles
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "addresses-file",
			Usage:    "Path to addresses.json file",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "chain-list-file",
			Usage:    "Path to chainList.json file",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "input-dir",
			Usage:    "Directory containing individual chain JSON files",
			Required: true,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func syncCodegenFiles(cliCtx *cli.Context) error {
	addressesFile := cliCtx.String("addresses-file")
	chainListFile := cliCtx.String("chain-list-file")
	inputDir := cliCtx.String("input-dir")

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))

	syncer, err := NewCodegenSyncer(lgr, addressesFile, chainListFile, inputDir)
	if err != nil {
		return err
	}
	lgr.Info("syncing codegen files")

	return syncer.Sync()
}
