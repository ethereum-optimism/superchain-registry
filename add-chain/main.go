package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/ethereum-optimism/optimism/op-e2e/bindings"
	"github.com/ethereum-optimism/superchain-registry/add-chain/cmd"
	"github.com/ethereum-optimism/superchain-registry/add-chain/config"
	"github.com/ethereum-optimism/superchain-registry/add-chain/flags"
	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:  "add-chain",
	Usage: "Add a new chain to the superchain-registry",
	Flags: []cli.Flag{
		flags.PublicRpcFlag,
		flags.SequencerRpcFlag,
		flags.ExplorerFlag,
		flags.SuperchainTargetFlag,
		flags.MonorepoDirFlag,
		flags.ChainNameFlag,
		flags.ChainShortNameFlag,
		flags.RollupConfigFlag,
		flags.DeploymentsDirFlag,
		flags.StandardChainCandidateFlag,
	},
	Action: entrypoint,
	Commands: []*cli.Command{
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
		fmt.Println("Loading .env")
		err = godotenv.Load()
	}

	if err != nil {
		panic("error loading .env file")
	}

	return app.Run(args)
}

func entrypoint(ctx *cli.Context) error {
	standardChainCandidate := ctx.Bool(flags.StandardChainCandidateFlag.Name)

	superchainLevel := superchain.Frontier // All chains enter as frontier chains

	publicRPC := ctx.String(flags.PublicRpcFlag.Name)
	sequencerRPC := ctx.String(flags.SequencerRpcFlag.Name)
	explorer := ctx.String(flags.ExplorerFlag.Name)
	superchainTarget := ctx.String(flags.SuperchainTargetFlag.Name)
	monorepoDir := ctx.String(flags.MonorepoDirFlag.Name)

	chainName := ctx.String(flags.ChainNameFlag.Name)
	rollupConfigPath := ctx.String(flags.RollupConfigFlag.Name)
	deploymentsDir := ctx.String(flags.DeploymentsDirFlag.Name)
	chainShortName := ctx.String(flags.ChainShortNameFlag.Name)
	if chainShortName == "" {
		return fmt.Errorf("must set chain-short-name (SCR_CHAIN_SHORT_NAME)")
	}

	// Get the current script filepath
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("error getting current filepath")
	}
	superchainRepoRoot := filepath.Dir(filepath.Dir(thisFile))

	fmt.Printf("Chain Name:                     %s\n", chainName)
	fmt.Printf("Chain Short Name:               %s\n", chainShortName)
	fmt.Printf("Superchain target:              %s\n", superchainTarget)
	fmt.Printf("Superchain-registry repo dir:   %s\n", superchainRepoRoot)
	fmt.Printf("Monorepo dir:                   %s\n", monorepoDir)
	fmt.Printf("Deployments directory:          %s\n", deploymentsDir)
	fmt.Printf("Rollup config filepath:         %s\n", rollupConfigPath)
	fmt.Printf("Public RPC endpoint:            %s\n", publicRPC)
	fmt.Printf("Sequencer RPC endpoint:         %s\n", sequencerRPC)
	fmt.Printf("Block Explorer:                 %s\n", explorer)
	fmt.Println()

	// Check if superchain target directory exists
	targetDir := filepath.Join(superchainRepoRoot, "superchain", "configs", superchainTarget)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("superchain target directory not found. Please follow instructions to add a superchain target in CONTRIBUTING.md: %s", targetDir)
	}

	l1RpcUrl, err := config.GetL1RpcUrl(superchainTarget)
	if err != nil {
		return fmt.Errorf("failed to retrieve L1 rpc url: %w", err)
	}

	addresses := make(map[string]string)
	err = readAddressesFromJSON(addresses, deploymentsDir)
	if err != nil {
		return fmt.Errorf("failed to read addresses from JSON files: %w", err)
	}

	isFaultProofs, err := inferIsFaultProofs(addresses["OptimismPortalProxy"], l1RpcUrl)
	if err != nil {
		return fmt.Errorf("failed to infer fault proofs status of chain: %w", err)
	}

	rollupConfig, err := config.ConstructChainConfig(rollupConfigPath, chainName, publicRPC, sequencerRPC, explorer, superchainLevel, standardChainCandidate)
	if err != nil {
		return fmt.Errorf("failed to construct rollup config: %w", err)
	}

	err = readAddressesFromChain(addresses, l1RpcUrl, isFaultProofs)
	if err != nil {
		return fmt.Errorf("failed to read addresses from chain: %w", err)
	}

	if rollupConfig.Plasma != nil {
		addresses["DAChallengeAddress"] = rollupConfig.Plasma.DAChallengeAddress.String()
	}

	addressList := superchain.AddressList{}
	err = mapToAddressList(addresses, &addressList)
	if err != nil {
		return fmt.Errorf("error converting map to AddressList: %w", err)
	}
	rollupConfig.Addresses = addressList

	l1RpcUrl, err = config.GetL1RpcUrl(superchainTarget)
	if err != nil {
		return fmt.Errorf("error getting l1RpcUrl: %w", err)
	}
	gpt, err := getGasPayingToken(l1RpcUrl, addressList.SystemConfigProxy)
	if err != nil {
		return fmt.Errorf("error inferring gas paying token: %w", err)
	}
	rollupConfig.GasPayingToken = gpt

	targetFilePath := filepath.Join(targetDir, chainShortName+".toml")
	err = config.WriteChainConfigTOML(rollupConfig, targetFilePath)
	if err != nil {
		return fmt.Errorf("error generating chain config .yaml file: %w", err)
	}

	fmt.Printf("Wrote config for new chain with identifier %s", rollupConfig.Identifier())
	return nil
}

func inferIsFaultProofs(optimismPortalProxyAddress, l1RpcUrl string) (bool, error) {
	// Portal version `3` is the first version of the `OptimismPortal` that supported the fault proof system.
	version, err := castCall(optimismPortalProxyAddress, "version()(string)", l1RpcUrl)
	if err != nil {
		return false, fmt.Errorf("failed to get OptimismPortalProxy.version(): %w", err)
	}

	version, err = strconv.Unquote(version)
	if err != nil {
		return false, fmt.Errorf("failed to parse OptimismPortalProxy.version(): %w", err)
	}
	majorVersion, err := strconv.ParseInt(strings.Split(version, ".")[0], 10, 32)
	if err != nil {
		return false, fmt.Errorf("failed to parse OptimismPortalProxy.version(): %w", err)
	}
	return majorVersion >= 3, nil
}

func getGasPayingToken(l1rpcURl string, SystemConfigAddress superchain.Address) (*superchain.Address, error) {
	client, err := ethclient.Dial(l1rpcURl)
	if err != nil {
		return nil, err
	}
	sc, err := bindings.NewSystemConfig(common.Address(SystemConfigAddress), client)
	if err != nil {
		return nil, err
	}

	opts := bind.CallOpts{}
	result, err := sc.GasPayingToken(&opts)

	if strings.Contains(err.Error(), "execution reverted") {
		// This happens when the SystemConfig contract
		// does not yet have the CGT functionality.
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if (result.Addr == common.Address{}) {
		// This happens with the SystemConfig contract
		// does have the CGT functionality, but it has
		// not been enabled.
		return nil, nil
	}

	return (*superchain.Address)(&result.Addr), nil
}
