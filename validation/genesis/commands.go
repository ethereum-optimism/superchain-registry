package genesis

import (
	"fmt"
	"strings"
)

type GeneratorFn func(uint64, string) string

// GenesisCreationCommand stores various functions which return a
// command, encoded as a string, which can be used to generate the
// Genesis.allocs object at some historical commit or range of commits in the
// github.com/ethereum-optimism/optimism repo. For example, the command
// may be an op-node subcommand invocation, or a Foundry script invocation.
// The invocation has changed over time, including the format of the inputs
// specified in the command line arguments.
var GenesisCreationCommand = map[string]GeneratorFn{
	"opnode1": opnode1,
	"opnode2": opnode2,
	"forge1":  forge1,
}

func opnode1(chainId uint64, l1rpcURL string) string { // runs from monorepo root
	return strings.Join([]string{
		"go run op-node/cmd/main.go genesis l2",
		fmt.Sprintf("--deploy-config=./packages/contracts-bedrock/deploy-config/%d.json", chainId),
		"--outfile.l2=expected-genesis.json",
		"--outfile.rollup=rollup.json",
		fmt.Sprintf("--deployment-dir=./packages/contracts-bedrock/deployments/%d", chainId),
		fmt.Sprintf("--l1-rpc=%s", l1rpcURL),
	},
		" ")
}

func opnode2(chainId uint64, l1rpcURL string) string { // runs from monorepo root
	return strings.Join([]string{
		"go run op-node/cmd/main.go genesis l2",
		fmt.Sprintf(" --deploy-config=./packages/contracts-bedrock/deploy-config/%d.json", chainId),
		"--outfile.l2=expected-genesis.json",
		"--outfile.rollup=rollup.json",
		fmt.Sprintf("--l1-deployments=./packages/contracts-bedrock/deployments/%d/.deploy", chainId),
		fmt.Sprintf("--l1-rpc=%s", l1rpcURL),
	},
		" ")
}

func forge1(chainId uint64, l1rpcURL string) string { // runs from packages/contracts-bedrock directory
	return strings.Join([]string{
		fmt.Sprintf("CONTRACT_ADDRESSES_PATH=./deployments/%d/.deploy", chainId),
		fmt.Sprintf("DEPLOY_CONFIG_PATH=./deploy-config/%d.json", chainId),
		"STATE_DUMP_PATH=statedump.json",
		"forge script ./scripts/L2Genesis.s.sol:L2Genesis",
		"--sig 'runWithStateDump()'",
	},
		" ")
}

var BuildCommand = map[string]string{
	"pnpm": "pnpm install --no-frozen-lockfile",
	"yarn": "yarn install --no-frozen-lockfile",
}
