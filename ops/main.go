package main

import (
	"fmt"
	"os"

	"github.com/ethereum-optimism/superchain-registry/ops/cmd"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:  "ops",
	Usage: "Utilities for working with the superchain-registry",
	Commands: []*cli.Command{
		&cmd.AddNewChainCmd,
		&cmd.PromoteToStandardCmd,
		&cmd.CheckRollupConfigCmd,
		&cmd.CompressGenesisCmd,
		&cmd.CheckGenesisCmd,
		&cmd.UpdateConfigsCmd,
	},
}

func main() {
	if err := runApp(os.Args); err != nil {
		fmt.Println(err)
		fmt.Println("*********************")
		fmt.Printf("FAILED: %s\n", app.Name)
		os.Exit(1)
	}
	fmt.Println("*********************")
	fmt.Printf("SUCCESS: %s\n", app.Name)
}

func runApp(args []string) error {
	// Load the appropriate .env file
	var err error
	if runningTests := os.Getenv("SCR_RUN_TESTS"); runningTests == "true" {
		fmt.Println("Loading .env.test")
		err = godotenv.Load("./testdata/.env.test")
	} else {
		err = godotenv.Load()
	}

	if err != nil && !os.IsNotExist(err) {
		fmt.Println(err)
		panic("error loading .env file")
	}

	return app.Run(args)
}
