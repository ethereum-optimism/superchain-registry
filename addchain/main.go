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
)

func main() {
	app := &cli.App{
		Name:   "add-chain",
		Usage:  "Add a new chain to the superchain-registry",
		Flags:  []cli.Flag{ChainTypeFlag},
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

	superchainLevel, err := getSuperchainLevel(chainType)
	if err != nil {
		return fmt.Errorf("failed to get superchain level: %w", err)
	}

	// Get the current script's directory
	superchainRepoPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	// Load environment variables
	viper.SetConfigName(".env")             // name of config file (without extension)
	viper.SetConfigType("env")              // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(superchainRepoPath) // path to look for the config file in
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	chainName := viper.GetString("CHAIN_NAME")
	superchainTarget := viper.GetString("SUPERCHAIN_TARGET")
	monorepoDir := viper.GetString("MONOREPO_DIR")
	deploymentsDir := viper.GetString("DEPLOYMENTS_DIR")
	rollupConfig := viper.GetString("ROLLUP_CONFIG")
	genesisConfig := viper.GetString("GENESIS_CONFIG")
	publicRPC := viper.GetString("PUBLIC_RPC")
	sequencerRPC := viper.GetString("SEQUENCER_RPC")
	explorer := viper.GetString("EXPLORER")

	fmt.Printf("Chain Name:                     %s\n", chainName)
	fmt.Printf("Superchain target:              %s\n", superchainTarget)
	fmt.Printf("Superchain-registry repo dir:   %s\n", superchainRepoPath)
	fmt.Printf("Reading from monrepo directory: %s\n", monorepoDir)
	fmt.Printf("With deployments directory:     %s\n", deploymentsDir)
	fmt.Printf("Rollup config:                  %s\n", rollupConfig)
	fmt.Printf("Genesis config:                 %s\n", genesisConfig)
	fmt.Printf("Public RPC endpoint:            %s\n", publicRPC)
	fmt.Printf("Sequencer RPC endpoint:         %s\n", sequencerRPC)
	fmt.Printf("Block Explorer:                 %s\n", explorer)
	fmt.Println()

	// Check if superchain target directory exists
	targetDir := filepath.Join(superchainRepoPath, "superchain", "configs", superchainTarget)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("superchain target directory not found. Please follow instructions to add a superchain target in CONTRIBUTING.md")
	}

	targetFilePath := filepath.Join(targetDir, chainName+".yaml")
	err = writeChainConfig(rollupConfig, targetFilePath, chainName, publicRPC, sequencerRPC, explorer, superchainLevel, superchainRepoPath, superchainTarget)
	if err != nil {
		return fmt.Errorf("error generating chain config .yaml file: %w", err)
	}

	contractAddresses := make(map[string]string)
	err = readAddressesFromJSON(contractAddresses, deploymentsDir)
	if err != nil {
		return fmt.Errorf("failed to read addresses from JSON files: %w", err)
	}

	l1RpcUrl, err := inferRpcUrl(superchainTarget)
	if err != nil {
		return fmt.Errorf("failed to infer rpc url: %w", err)
	}

	err = readAddressesFromChain(contractAddresses, l1RpcUrl)
	if err != nil {
		return fmt.Errorf("failed to read addresses from chain: %w", err)
	}

	err = writeAddressesToJSON(contractAddresses, superchainRepoPath, superchainTarget, chainName)
	if err != nil {
		return fmt.Errorf("failed to write contract addresses to JSON file: %w", err)
	}

	err = createExtraGenesisData(superchainRepoPath, superchainTarget, monorepoDir, genesisConfig, chainName)
	if err != nil {
		return fmt.Errorf("failed to create extra genesis data: %w", err)
	}
	return nil
}
