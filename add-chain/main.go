package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/optimism/op-e2e/bindings"
	"github.com/ethereum-optimism/superchain-registry/add-chain/cmd"
	"github.com/ethereum-optimism/superchain-registry/add-chain/config"
	"github.com/ethereum-optimism/superchain-registry/add-chain/flags"
	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/genesis"
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
		flags.GenesisFlag,
		flags.DeploymentsDirFlag,
		flags.StandardChainCandidateFlag,
		flags.GenesisCreationCommit,
		flags.DeployConfigFlag,
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
	genesisPath := ctx.String(flags.GenesisFlag.Name)
	deployConfigPath := ctx.String(flags.DeployConfigFlag.Name)
	genesisCreationCommit := ctx.String(flags.GenesisCreationCommit.Name)
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
	fmt.Printf("Genesis filepath:               %s\n", genesisPath)
	fmt.Printf("Deploy config filepath:         %s\n", deployConfigPath)
	fmt.Printf("Genesis creation commit:        %s\n", genesisCreationCommit)
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

	var addresses superchain.AddressList
	err = readAddressesFromJSON(&addresses, deploymentsDir)
	if err != nil {
		return fmt.Errorf("failed to read addresses from JSON files: %w", err)
	}

	isFaultProofs, err := inferIsFaultProofs(addresses.SystemConfigProxy, addresses.OptimismPortalProxy, l1RpcUrl)
	if err != nil {
		return fmt.Errorf("failed to infer fault proofs status of chain: %w", err)
	}

	rollupConfig, err := config.ConstructChainConfig(rollupConfigPath, genesisPath, chainName, publicRPC, sequencerRPC, explorer, superchainLevel, standardChainCandidate)
	if err != nil {
		return fmt.Errorf("failed to construct rollup config: %w", err)
	}

	err = readAddressesFromChain(&addresses, l1RpcUrl, isFaultProofs)
	if err != nil {
		return fmt.Errorf("failed to read addresses from chain: %w", err)
	}

	if rollupConfig.AltDA != nil {
		addresses.DAChallengeAddress = *rollupConfig.AltDA.DAChallengeAddress
	}

	rollupConfig.Addresses = addresses

	l1RpcUrl, err = config.GetL1RpcUrl(superchainTarget)
	if err != nil {
		return fmt.Errorf("error getting l1RpcUrl: %w", err)
	}
	gpt, err := getGasPayingToken(l1RpcUrl, addresses.SystemConfigProxy)
	if err != nil {
		return fmt.Errorf("error inferring gas paying token: %w", err)
	}
	rollupConfig.GasPayingToken = gpt

	targetFilePath := filepath.Join(targetDir, chainShortName+".toml")
	err = config.WriteChainConfigTOML(rollupConfig, targetFilePath)
	if err != nil {
		return fmt.Errorf("error generating chain config .yaml file: %w", err)
	}

	fmt.Printf("✅ Wrote config for new chain with identifier %s", rollupConfig.Identifier())

	folderName := fmt.Sprintf("%d", rollupConfig.ChainID)
	if runningTests := os.Getenv("SCR_RUN_TESTS"); runningTests == "true" {
		folderName = folderName + "-test"
	}
	genesisValidationInputsDir := filepath.Join(superchainRepoRoot, "validation", "genesis", "validation-inputs", folderName)
	err = os.MkdirAll(genesisValidationInputsDir, os.ModePerm)
	if err != nil {
		return err
	}
	err = copyDeployConfigFile(deployConfigPath, genesisValidationInputsDir)
	if err != nil {
		return fmt.Errorf("error copying deploy-config json file: %w", err)
	}
	fmt.Printf("✅ Copied deploy-config json file to validation module")

	err = writeGenesisValidationMetadata(genesisCreationCommit, genesisValidationInputsDir)
	if err != nil {
		return fmt.Errorf("error writing genesis validation metadata file: %w", err)
	}
	fmt.Printf("✅ Wrote genesis validation metadata file")

	return nil
}

func inferIsFaultProofs(systemConfigProxyAddress, optimismPortalProxyAddress superchain.Address, l1RpcUrl string) (bool, error) {
	tokenAddress, err := getGasPayingToken(l1RpcUrl, systemConfigProxyAddress)
	if err != nil {
		return false, fmt.Errorf("failed to query for gasPayingToken: %w", err)
	}
	if tokenAddress != nil {
		return false, nil
	}

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
	if err != nil {
		if strings.Contains(err.Error(), "execution reverted") {
			// This happens when the SystemConfig contract
			// does not yet have the CGT functionality.
			return nil, nil
		}
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

func copyDeployConfigFile(sourcePath string, targetDir string) error {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(targetDir, "deploy-config.json"), data, os.ModePerm)
}

func writeGenesisValidationMetadata(commit string, targetDir string) error {
	// Define default metadata params:
	// These may not be sufficient to make the genesis validation work,
	// but we address that with some manual trial-and-error intervention
	// involving OPLabs engineers after the add-chain command runs.
	const defaultNodeVersion = "18.12.1"
	const defaultMonorepoBuildCommand = "pnpm"
	const defaultGenesisCreationCommand = "opnode2" // See validation/genesis/commands.go
	vm := genesis.ValidationMetadata{
		GenesisCreationCommit:  commit,
		NodeVersion:            defaultNodeVersion,
		MonorepoBuildCommand:   defaultMonorepoBuildCommand,
		GenesisCreationCommand: defaultGenesisCreationCommand,
	}
	data, err := toml.Marshal(vm)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(targetDir, "meta.toml"), data, os.ModePerm)
}
