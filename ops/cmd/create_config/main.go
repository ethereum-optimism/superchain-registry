package main

import (
	"fmt"
	"os"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/urfave/cli/v2"
)

var (
	StateFilename = &cli.StringFlag{
		Name:      "state-filename",
		Usage:     "Filename of an op-deployer state file.",
		Required:  true,
		TakesFile: true,
	}
	Shortname = &cli.StringFlag{
		Name:     "shortname",
		Usage:    "Shortname of the chain.",
		Required: true,
		Value:    "newchain",
	}
)

func main() {
	app := &cli.App{
		Name:  "create-config",
		Usage: "Turns a state file into a chain config in the stating directory.",
		Flags: []cli.Flag{
			StateFilename,
			Shortname,
		},
		Action: action,
	}
	if err := app.Run(os.Args); err != nil {
		output.WriteStderr("%v", err)
		os.Exit(1)
	}
}

func action(cliCtx *cli.Context) error {
	wd, err := paths.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	statePath := cliCtx.String(StateFilename.Name)

	err = manage.GenerateChainArtifacts(statePath, wd, cliCtx.String(Shortname.Name), nil, nil, 0, "")
	if err != nil {
		return fmt.Errorf("failed to generate chain config: %w", err)
	}

	output.WriteOK("done")
	return nil
}
