package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ethereum-optimism/superchain-registry/add-chain/cmd"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:   "add-chain",
	Usage:  "Utilities for working with the superchain-registry",
	Action: entrypoint,
	Commands: []*cli.Command{
		&cmd.AddChainCmd,
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

// call default command (AddChainCmd) if no subcommand is provided
func entrypoint(c *cli.Context) error {
	if c.Args().Present() {
		// Unknown command or arguments were provided
		return cli.ShowAppHelp(c)
	}

	// No subcommand provided; invoke the default command

	// Create a new flag set for the default command
	set := flag.NewFlagSet(cmd.AddChainCmd.Name, flag.ExitOnError)
	for _, f := range cmd.AddChainCmd.Flags {
		_ = f.Apply(set)
	}

	// Create a new context with the new flag set
	ctx := cli.NewContext(c.App, set, c)

	// Set the command to the default command
	ctx.Command = &cmd.AddChainCmd

	// Invoke the default command's Run method
	return cmd.AddChainCmd.Run(ctx)
}
