package main

import (
	"fmt"
	"os"
	"path"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"

	"github.com/urfave/cli/v2"
)

var (
	StatePath = &cli.StringFlag{
		Name:      "state-filename",
		Usage:     "Path to an op-deployer state.json file.",
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
			StatePath,
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

	statePath := cliCtx.String(StatePath.Name)
	output.WriteStderr("reading state file from %s", statePath)
	var st state.State
	if err := paths.ReadJSONFile(statePath, &st); err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	type manifest struct {
		Name string `yaml:"name"`
		L2   struct {
			Chains []struct {
				Name    string `yaml:"name"`
				ChainID int64  `yaml:"chain_id"`
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
		if m.L2.Chains[i].ChainID != int64(st.AppliedIntent.Chains[i].ID.Big().Int64()) {
			return fmt.Errorf("chain ID mismatch for chain at index %d : manifest %d, state %d",
				i,
				m.L2.Chains[i].ChainID, st.AppliedIntent.Chains[i].ID.Big().Int64())
		}
		err = manage.GenerateChainArtifacts(statePath, wd, m.L2.Chains[i].Name, &m.L2.Chains[i].Name, &m.Name, i)
		if err != nil {
			return fmt.Errorf("failed to generate chain config: %w", err)
		}
	}

	output.WriteOK("writing superchain definition file")

	sD := config.SuperchainDefinition{
		Name:                   m.Name,
		ProtocolVersionsAddr:   config.NewChecksummedAddress(st.SuperchainDeployment.ProtocolVersionsProxyAddress),
		SuperchainConfigAddr:   config.NewChecksummedAddress(st.SuperchainDeployment.SuperchainConfigProxyAddress),
		OPContractsManagerAddr: config.NewChecksummedAddress(st.ImplementationsDeployment.OpcmAddress),
		Hardforks:              config.Hardforks{}, // superchain wide hardforks are added after chains are in the registry.
		L1: config.SuperchainL1{
			ChainID: st.AppliedIntent.L1ChainID,
		},
	}

	// Validation and conflict resolution will be handled
	// by the sync staging command.
	err = paths.WriteTOMLFile(path.Join(paths.StagingDir(wd), "superchain.toml"), sD)
	if err != nil {
		return fmt.Errorf("failed to write superchain definition file: %w", err)
	}

	output.WriteOK("done")
	return nil
}
