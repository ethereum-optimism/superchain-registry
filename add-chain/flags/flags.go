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
	ChainNameFlag = &cli.StringFlag{
		Name:     "chain-name",
		Value:    "",
		EnvVars:  prefixEnvVars("CHAIN_NAME"),
		Usage:    "Custom name of the chain",
		Required: true,
	}
	ChainShortNameFlag = &cli.StringFlag{
		Name:     "chain-short-name",
		Value:    "",
		EnvVars:  prefixEnvVars("CHAIN_SHORT_NAME"),
		Usage:    "Custom short name of the chain",
		Required: true,
	}
	PublicRpcFlag = &cli.StringFlag{
		Name:     "public-rpc",
		Value:    "",
		EnvVars:  prefixEnvVars("PUBLIC_RPC"),
		Usage:    "L2 node public rpc url",
		Required: true,
	}
	SequencerRpcFlag = &cli.StringFlag{
		Name:     "sequencer-rpc",
		Value:    "",
		EnvVars:  prefixEnvVars("SEQUENCER_RPC"),
		Usage:    "sequencer rpc url",
		Required: true,
	}
	ExplorerFlag = &cli.StringFlag{
		Name:     "explorer",
		Value:    "",
		EnvVars:  prefixEnvVars("EXPLORER"),
		Usage:    "block explorer url",
		Required: true,
	}
	SuperchainTargetFlag = &cli.StringFlag{
		Name:     "superchain-target",
		Value:    "",
		EnvVars:  prefixEnvVars("SUPERCHAIN_TARGET"),
		Usage:    "superchain this L2 will belong to (mainnet or sepolia)",
		Required: true,
	}
	MonorepoDirFlag = &cli.StringFlag{
		Name:     "monorepo-dir",
		Value:    "",
		EnvVars:  prefixEnvVars("MONOREPO_DIR"),
		Usage:    "path to local 'ethereum-optimism/optimism' monorepo",
		Required: false,
	}
	RollupConfigFlag = &cli.StringFlag{
		Name:     "rollup-config",
		EnvVars:  prefixEnvVars("ROLLUP_CONFIG"),
		Usage:    "Filepath to rollup.json input file",
		Required: true,
	}
	GenesisFlag = &cli.StringFlag{
		Name:     "genesis",
		EnvVars:  prefixEnvVars("GENESIS"),
		Usage:    "Filepath to genesis.json input file",
		Required: true,
	}
	DeploymentsDirFlag = &cli.StringFlag{
		Name:     "deployments-dir",
		Value:    "",
		EnvVars:  prefixEnvVars("DEPLOYMENTS_DIR"),
		Usage:    "Directory containing L1 Contract deployment addresses",
		Required: true,
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
		Usage:    "globally unique ID of chain",
		Required: true,
	}
)

var L2GenesisHeaderFlag = &cli.PathFlag{
	Name:    "l2-genesis-header",
	Value:   "genesis-header.json",
	Usage:   "Alternative to l2-genesis flag, if genesis-state is omitted. Path to block header at genesis",
	EnvVars: prefixEnvVars("L2_GENESIS_HEADER"),
}
