package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func main() {
	chainType := flag.String("chain_type", "", "Type of chain to add. Must be 'standard' or 'frontier'.")
	help := flag.Bool("help", false, "Show help information")
	flag.Parse()

	if *help {
		showUsage()
		return
	}

	superchainLevel, err := getSuperchainLevel(*chainType)
	if err != nil {
		fmt.Println("Failed to get superchain level:", err)
		showUsage()
		os.Exit(1)
	}

	// Get the current script's directory
	scriptDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		os.Exit(1)
	}
	superchainRepoPath := filepath.Dir(scriptDir) // Parent directory

	// Load environment variables
	viper.SetConfigName(".env")             // name of config file (without extension)
	viper.SetConfigType("env")              // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(superchainRepoPath) // path to look for the config file in
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error reading config file:", err)
		os.Exit(1)
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
	fmt.Printf("Reading from monrepo directory: %s\n", monorepoDir)
	fmt.Printf("With deployments directory:     %s\n", deploymentsDir)
	fmt.Printf("Rollup config:                  %s\n", rollupConfig)
	fmt.Printf("Genesis config:                 %s\n", genesisConfig)
	fmt.Printf("Public RPC endpoint:            %s\n", publicRPC)
	fmt.Printf("Sequencer RPC endpoint:         %s\n", sequencerRPC)
	fmt.Printf("Block Explorer:                 %s\n", explorer)

	// Check if superchain target directory exists
	targetDir := filepath.Join(superchainRepoPath, "superchain", "configs", superchainTarget)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Println("Superchain target directory not found. Please follow instructions to add a superchain target in CONTRIBUTING.md")
		os.Exit(1)
	}

	inputFilePath := filepath.Join(deploymentsDir, "RollupConfig.json")
	targetFilePath := filepath.Join(targetDir, chainName+".yaml")
	err = writeChainConfig(inputFilePath, targetFilePath, chainName, publicRPC, sequencerRPC, superchainLevel, superchainRepoPath, superchainTarget)
	if err != nil {
		fmt.Println("Error generating chain config .yaml file:", err)
		os.Exit(1)
	}

	var contractAddresses map[string]string
	err = readAddressesFromJSON(contractAddresses, deploymentsDir)
	if err != nil {
		fmt.Printf("Failed to read addresses from JSON files: %v\n", err)
		os.Exit(1)
	}

	l1RpcUrl, err := inferRpcUrl(superchainTarget)
	if err != nil {
		fmt.Printf("Failed to infer rpc url: %v\n", err)
		os.Exit(1)
	}

	err = readAddressesFromChain(contractAddresses, l1RpcUrl)
	if err != nil {
		fmt.Printf("Failed to read addresses from chain: %v\n", err)
		os.Exit(1)
	}

	err = writeAddressesToJSON(contractAddresses, superchainRepoPath, superchainTarget, chainName)
	if err != nil {
		fmt.Printf("Failed to write contract addresses to JSON file: %v\n", err)
		os.Exit(1)
	}

	err = createExtraGenesisData(superchainRepoPath, superchainTarget, monorepoDir, genesisConfig, chainName)
	if err != nil {
		fmt.Printf("Failed to create extra genesis data: %v\n", err)
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Println("Usage: command <chain_type> [-h|--help]")
	fmt.Println("  chain_type: The type of chain to add. Must be 'standard' or 'frontier'.")
	fmt.Println("  -h, --help: Show this usage information.")
}
