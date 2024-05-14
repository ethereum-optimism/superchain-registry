package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

var (
	ChainTypeFlag = &cli.StringFlag{
		Name:     "chain-type",
		Value:    "",
		Usage:    "Type of chain (either standard or frontier)",
		Required: true,
	}
	ChainNameFlag = &cli.StringFlag{
		Name:     "chain-name",
		Value:    "",
		Usage:    "Custom name of the chain",
		Required: false,
	}
	RollupConfigFlag = &cli.StringFlag{
		Name:     "rollup-config",
		Value:    "",
		Usage:    "Filepath to rollup.json input file",
		Required: false,
	}
	TestFlag = &cli.BoolFlag{
		Name:     "test",
		Value:    false,
		Usage:    "Indicates if go tests are being run",
		Required: false,
	}
)

func main() {
	app := &cli.App{
		Name:   "add-chain",
		Usage:  "Add a new chain to the superchain-registry",
		Flags:  []cli.Flag{ChainTypeFlag, ChainNameFlag, RollupConfigFlag, TestFlag},
		Action: entrypoint,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		fmt.Println("*********************")
		fmt.Printf("FAILED: %s\n", app.Name)
		os.Exit(1)
	}
	fmt.Println("*********************")
	fmt.Printf("SUCCESS: %s\n", app.Name)
}

func entrypoint(ctx *cli.Context) error {
	chainType := ctx.String(ChainTypeFlag.Name)
	runningTests := ctx.Bool(TestFlag.Name)

	superchainLevel, err := getSuperchainLevel(chainType)
	if err != nil {
		return fmt.Errorf("failed to get superchain level: %w", err)
	}

	// Get the current script's directory
	superchainRepoPath, err := os.Getwd()
	envFilename := ".env"
	envPath := "."
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}
	if runningTests {
		envFilename = ".env.test"
		envPath = "./testdata"
		superchainRepoPath = filepath.Join(superchainRepoPath, "testdata")
	}

	// Load environment variables
	viper.SetConfigName(envFilename) // name of config file (without extension)
	viper.SetConfigType("env")       // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(envPath)     // path to look for the config file in
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	publicRPC := viper.GetString("PUBLIC_RPC")
	sequencerRPC := viper.GetString("SEQUENCER_RPC")
	explorer := viper.GetString("EXPLORER")
	superchainTarget := viper.GetString("SUPERCHAIN_TARGET")
	deploymentsDir := viper.GetString("DEPLOYMENTS_DIR")
	chainName := viper.GetString("CHAIN_NAME")

	// Allow cli flags to override env vars
	if ctx.IsSet("chain-name") {
		chainName = ctx.String("chain-name")
	}
	rollupConfigPath := viper.GetString("ROLLUP_CONFIG")
	if ctx.IsSet("rollup-config") {
		rollupConfigPath = ctx.String("rollup-config")
	}

	fmt.Printf("Chain Name:                     %s\n", chainName)
	fmt.Printf("Superchain target:              %s\n", superchainTarget)
	fmt.Printf("Superchain-registry repo dir:   %s\n", superchainRepoPath)
	fmt.Printf("With deployments directory:     %s\n", deploymentsDir)
	fmt.Printf("Rollup config filepath:         %s\n", rollupConfigPath)
	fmt.Printf("Public RPC endpoint:            %s\n", publicRPC)
	fmt.Printf("Sequencer RPC endpoint:         %s\n", sequencerRPC)
	fmt.Printf("Block Explorer:                 %s\n", explorer)
	fmt.Println()

	// Check if superchain target directory exists
	targetDir := filepath.Join(superchainRepoPath, "superchain", "configs", superchainTarget)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("superchain target directory not found. Please follow instructions to add a superchain target in CONTRIBUTING.md")
	}

	rollupConfig, err := constructChainConfig(rollupConfigPath, chainName, publicRPC, sequencerRPC, explorer, superchainLevel)
	if err != nil {
		return fmt.Errorf("failed to construct rollup config: %w", err)
	}
	contractAddresses := make(map[string]string)
	if rollupConfig.Plasma != nil {
		// Store this address before it gets removed from rollupConfig
		contractAddresses["DAChallengeAddress"] = rollupConfig.Plasma.DAChallengeAddress.String()
	}

	targetFilePath := filepath.Join(targetDir, chainName+".yaml")
	err = writeChainConfig(rollupConfig, targetFilePath, superchainRepoPath, superchainTarget)
	if err != nil {
		return fmt.Errorf("error generating chain config .yaml file: %w", err)
	}

	err = readAddressesFromJSON(contractAddresses, deploymentsDir)
	if err != nil {
		return fmt.Errorf("failed to read addresses from JSON files: %w", err)
	}

	l1RpcUrl, err := getL1RpcUrl(superchainTarget)
	if err != nil {
		return fmt.Errorf("failed to retrieve L1 rpc url: %w", err)
	}

	err = readAddressesFromChain(contractAddresses, l1RpcUrl)
	if err != nil {
		return fmt.Errorf("failed to read addresses from chain: %w", err)
	}

	err = writeAddressesToJSON(contractAddresses, superchainRepoPath, superchainTarget, chainName)
	if err != nil {
		return fmt.Errorf("failed to write contract addresses to JSON file: %w", err)
	}

	return nil
}
