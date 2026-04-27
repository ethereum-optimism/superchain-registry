package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/report"

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
	OpDeployerVersion = &cli.StringFlag{
		Name:     "op-deployer-version",
		Usage:    "Version of op-deployer used to deploy the chain(s). If not provided, the version will be inferred from the state file.",
		Required: false,
	}
	OpDeployerBinDir = &cli.StringFlag{
		Name:    "op-deployer-bin-dir",
		Usage:   "Path to the directory containing op-deployer binaries.",
		EnvVars: []string{"DEPLOYER_CACHE_DIR"},
		Value:   defaultBinDir(),
	}
	L1ContractsVersion = &cli.StringFlag{
		Name:    "l1-contracts-version",
		Usage:   "Version tag of the L1 contracts (e.g., 'op-contracts/v4.0.0'). If not specified, will be auto-detected from OPCM address.",
		EnvVars: []string{"L1_CONTRACTS_VERSION"},
		Value:   "",
	}
)

func main() {
	app := &cli.App{
		Name:  "create-config",
		Usage: "Turns an op-deployer state.json file and a devnet manifest.yaml file into multiple chain configs and a superchain manifest file in the staging directory.",
		Flags: []cli.Flag{
			StatePath,
			ManifestPath,
			OpDeployerVersion,
			OpDeployerBinDir,
			L1ContractsVersion,
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

	l1ContractsVersion := cliCtx.String(L1ContractsVersion.Name)
	if l1ContractsVersion == "" {
		var err error
		l1ContractsVersion, err = report.GetContractsReleaseForOpcm(statePath)
		if err != nil {
			return fmt.Errorf("failed to determine L1 contracts version: %w", err)
		}
	}
	output.WriteStderr("L1 contracts version: %s", l1ContractsVersion)

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

	st, err := deployer.ReadOpaqueStateFile(statePath)
	if err != nil {
		return fmt.Errorf("failed to read opaque state file: %w", err)
	}

	numChains, err := st.GetNumChains()
	if err != nil {
		return fmt.Errorf("failed to read number of chains: %w", err)
	}
	if numChains != len(m.L2.Chains) {
		return fmt.Errorf("number of chains in manifest file (%d) does not match number of chains in state file (%d)", len(m.L2.Chains), numChains)
	}

	output.WriteOK("inflating chain configs")
	opDeployerVersion := cliCtx.String(OpDeployerVersion.Name)
	opDeployerBinDir := cliCtx.String(OpDeployerBinDir.Name)

	output.WriteWarn("⚠️  Config generation behavior has changed: now generates only essential addresses by default.")
	output.WriteWarn("📄 All addresses are still available in addresses.json")

	for i := 0; i < numChains; i++ {
		chainID, err := st.GetChainID(i)
		if err != nil {
			return fmt.Errorf("failed to read chain id: %w", err)
		}
		chain := m.L2.Chains[i]
		if chain.ChainID != int64(chainID) {
			return fmt.Errorf("chain ID mismatch for chain at index %d : manifest %d, state %d",
				i,
				chain.ChainID, chainID)
		}
		if err := manage.GenerateChainArtifacts(
			statePath,
			wd,
			chain.Name,
			&chain.Name,
			&m.Name,
			i,
			opDeployerVersion,
			opDeployerBinDir,
			l1ContractsVersion,
		); err != nil {
			return fmt.Errorf("failed to generate chain config: %w", err)
		}
	}

	output.WriteOK("writing superchain definition file")

	sD, err := manage.InflateSuperchainDefinition(m.Name, st)
	if err != nil {
		return fmt.Errorf("failed to inflate superchain definition: %w", err)
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

func defaultBinDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get home directory: %v", err))
	}

	return filepath.Join(homeDir, ".cache", "op-deployer")
}
