package flags

import (
	"github.com/urfave/cli/v2"
)

const EnvVarPrefix = "SCR"

func prefixEnvVars(names ...string) []string {
	envs := make([]string, 0, len(names))
	for _, name := range names {
		envs = append(envs, EnvVarPrefix+"_"+name)
	}
	return envs
}

var (
	ChainTypeFlag = &cli.StringFlag{
		Name:     "chain-type",
		Value:    "frontier",
		EnvVars:  prefixEnvVars("CHAIN_TYPE"),
		Usage:    "Type of chain (either standard or frontier)",
		Required: false,
	}
	ChainNameFlag = &cli.StringFlag{
		Name:     "chain-name",
		Value:    "",
		EnvVars:  prefixEnvVars("CHAIN_NAME"),
		Usage:    "Custom name of the chain",
		Required: false,
	}
	ChainShortNameFlag = &cli.StringFlag{
		Name:     "chain-short-name",
		Value:    "",
		EnvVars:  prefixEnvVars("CHAIN_SHORT_NAME"),
		Usage:    "Custom short name of the chain",
		Required: false,
	}
	RollupConfigFlag = &cli.StringFlag{
		Name:     "rollup-config",
		EnvVars:  prefixEnvVars("ROLLUP_CONFIG"),
		Usage:    "Filepath to rollup.json input file",
		Required: true,
	}
	DeploymentsDirFlag = &cli.StringFlag{
		Name:     "deployments-dir",
		Value:    "",
		EnvVars:  prefixEnvVars("DEPLOYMENTS_DIR"),
		Usage:    "Directory containing L1 Contract deployment addresses",
		Required: false,
	}
	StandardChainCandidateFlag = &cli.BoolFlag{
		Name:     "standard-chain-candidate",
		Value:    false,
		EnvVars:  prefixEnvVars("STANDARD_CHAIN_CANDIDATE"),
		Usage:    "Whether the chain is a candidate to become a standard chain. Will be subject to most standard chain validation checks",
		Required: false,
	}
	ChainIdFlag = &cli.Uint64Flag{
		Name:     "chain-id",
		Usage:    "ID of chain to promote",
		Required: true,
	}
)
