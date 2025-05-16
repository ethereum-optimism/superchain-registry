package main

import (
	"fmt"
	"os"
	"path"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/inspect"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
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
	ManifestPath = &cli.StringFlag{
		Name:      "manifest-path",
		Usage:     "Path to a manifest.yaml file specifying chain names.",
		Required:  true,
		TakesFile: true,
	}
)

func main() {
	app := &cli.App{
		Name:  "create-config",
		Usage: "Turns an op-deployer state.json file and a devnet manifest.yaml file into multiple chain configs and a superchain manifest file in the staging directory.",
		Flags: []cli.Flag{
			StateFilename,
			ManifestPath,
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
	output.WriteStderr("reading state file from %s", statePath)
	var st state.State
	if err := paths.ReadJSONFile(statePath, &st); err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	type manifest struct {
		Name string `yaml:"name"`
		L2   struct {
			Chains []struct {
				Name string `yaml:"name"`
			} `yaml:"chains"`
		} `yaml:"l2"`
	}
	var m manifest
	if err := paths.ReadYAMLFile(cliCtx.String(ManifestPath.Name), &m); err != nil {
		return fmt.Errorf("failed to read manifest file: %w", err)
	}

	if len(m.L2.Chains) != len(st.AppliedIntent.Chains) {
		return fmt.Errorf(
			"number of chains in manifest file (%d) does not match number of chains in state file (%d)",
			len(m.L2.Chains), len(st.AppliedIntent.Chains))
	}

	output.WriteOK("inflating chain configs")
	for i := 0; i < len(st.AppliedIntent.Chains); i++ {
		cfg, err := manage.InflateChainConfig(&st, i)
		if err != nil {
			return fmt.Errorf("failed to inflate chain config: %w", err)
		}

		// just use devnet name with numerical suffix
		// OR parse the manifest.yaml file and get the name from there
		cfg.ShortName = m.L2.Chains[i].Name
		cfg.Name = m.L2.Chains[i].Name

		cfg.Superchain = config.Superchain(m.Name) // Each devnet forms its own unique superchain

		output.WriteOK("reading genesis")
		genesis, _, err := inspect.GenesisAndRollup(&st, st.AppliedIntent.Chains[i].ID)
		if err != nil {
			return fmt.Errorf("failed to get genesis: %w", err)
		}

		stagingDir := paths.StagingDir(wd)

		output.WriteOK("writing chain config")
		if err := paths.WriteTOMLFile(path.Join(stagingDir, cfg.ShortName+".toml"), cfg); err != nil {
			return fmt.Errorf("failed to write chain config: %w", err)
		}

		output.WriteOK("writing genesis")
		if err := manage.WriteGenesis(wd, path.Join(stagingDir, cfg.ShortName+".json.zst"), genesis); err != nil {
			return fmt.Errorf("failed to write genesis: %w", err)
		}
	}

	output.WriteOK("writing superchain manifest")

	sD := config.SuperchainDefinition{
		Name:                   m.Name,
		ProtocolVersionsAddr:   config.NewChecksummedAddress(st.SuperchainDeployment.ProtocolVersionsProxyAddress),
		SuperchainConfigAddr:   config.NewChecksummedAddress(st.SuperchainDeployment.SuperchainConfigProxyAddress),
		OPContractsManagerAddr: config.NewChecksummedAddress(st.ImplementationsDeployment.OpcmAddress),
		Hardforks:              config.Hardforks{}, // superchain wide hardforks are added after chains are in the registry.
		L1: config.SuperchainL1{
			ChainID: st.AppliedIntent.L1ChainID,
			// TODO grab this from a lookup or prompt user to fill it out afterwards
		},
	}

	// We write this to {m.Name}.superchain-toml and when it gets sync'ed later
	// it will be moved to the appropriate superchain directory and renamed
	// to superchain.toml. Validation and conflict resolution will be handled
	// by the sync staging command.
	err = paths.WriteTOMLFile(path.Join(paths.StagingDir(wd), m.Name+".superchain-toml"), sD)
	if err != nil {
		return fmt.Errorf("failed to write superchain manifest: %w", err)
	}

	output.WriteOK("done")
	return nil
}
