package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:     "add-chain",
	Usage:    "Add a new chain to the superchain-registry",
	Flags:    []cli.Flag{ChainTypeFlag, ChainNameFlag, ChainShortNameFlag, RollupConfigFlag, DeploymentsDirFlag, TestFlag, StandardChainCandidateFlag},
	Action:   entrypoint,
	Commands: []*cli.Command{&PromoteToStandardCmd, &CheckRollupConfigCmd},
}

var (
	ChainTypeFlag = &cli.StringFlag{
		Name:     "chain-type",
		Value:    "frontier",
		Usage:    "Type of chain (either standard or frontier)",
		Required: false,
	}
	ChainNameFlag = &cli.StringFlag{
		Name:     "chain-name",
		Value:    "",
		Usage:    "Custom name of the chain",
		Required: false,
	}
	ChainShortNameFlag = &cli.StringFlag{
		Name:     "chain-short-name",
		Value:    "",
		Usage:    "Custom short name of the chain",
		Required: false,
	}
	RollupConfigFlag = &cli.StringFlag{
		Name:     "rollup-config",
		Value:    "",
		Usage:    "Filepath to rollup.json input file",
		Required: false,
	}
	DeploymentsDirFlag = &cli.StringFlag{
		Name:     "deployments-dir",
		Value:    "",
		Usage:    "Directory containing L1 Contract deployment addresses",
		Required: false,
	}
	TestFlag = &cli.BoolFlag{
		Name:     "test",
		Value:    false,
		Usage:    "Indicates if go tests are being run",
		Required: false,
	}
	StandardChainCandidateFlag = &cli.BoolFlag{
		Name:     "standard-chain-candidate",
		Value:    false,
		Usage:    "Whether the chain is a candidate to become a standard chain. Will be subject to most standard chain validation checks",
		Required: false,
	}
)

func main() {
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
	standardChainCandidate := ctx.Bool(StandardChainCandidateFlag.Name)

	if standardChainCandidate && chainType == "standard" {
		return errors.New("cannot set both chainType=standard and standard-chain-candidate=true")
	}

	superchainLevel, err := getSuperchainLevel(chainType)
	if err != nil {
		return fmt.Errorf("failed to get superchain level: %w", err)
	}

	// Get the current script's directory
	superchainRepoReadPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}
	superchainRepoWritePath := filepath.Dir(superchainRepoReadPath)
	envFilename := ".env"
	envPath := "."
	if runningTests {
		envFilename = ".env.test"
		envPath = "./testdata"
		superchainRepoReadPath = filepath.Join(superchainRepoReadPath, "testdata")
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
	chainName := viper.GetString("CHAIN_NAME")
	chainShortName := viper.GetString("CHAIN_SHORT_NAME")

	// Allow cli flags to override env vars
	if ctx.IsSet("chain-name") {
		chainName = ctx.String("chain-name")
	}
	if ctx.IsSet("chain-short-name") {
		chainShortName = ctx.String("chain-short-name")
	}
	rollupConfigPath := viper.GetString("ROLLUP_CONFIG")
	if ctx.IsSet("rollup-config") {
		rollupConfigPath = ctx.String("rollup-config")
	}
	deploymentsDir := viper.GetString("DEPLOYMENTS_DIR")
	if ctx.IsSet(DeploymentsDirFlag.Name) {
		deploymentsDir = ctx.String(DeploymentsDirFlag.Name)
	}

	if chainShortName == "" {
		return fmt.Errorf("must set chain-short-name (CHAIN_SHORT_NAME)")
	}

	fmt.Printf("Chain Name:                     %s\n", chainName)
	fmt.Printf("Chain Short Name:               %s\n", chainShortName)
	fmt.Printf("Superchain target:              %s\n", superchainTarget)
	fmt.Printf("Superchain-registry repo dir:   %s\n", superchainRepoReadPath)
	fmt.Printf("With deployments directory:     %s\n", deploymentsDir)
	fmt.Printf("Rollup config filepath:         %s\n", rollupConfigPath)
	fmt.Printf("Public RPC endpoint:            %s\n", publicRPC)
	fmt.Printf("Sequencer RPC endpoint:         %s\n", sequencerRPC)
	fmt.Printf("Block Explorer:                 %s\n", explorer)
	fmt.Println()

	// Check if superchain target directory exists
	targetDir := filepath.Join(superchainRepoWritePath, "superchain", "configs", superchainTarget)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("superchain target directory not found. Please follow instructions to add a superchain target in CONTRIBUTING.md: %s", targetDir)
	}

	l1RpcUrl, err := getL1RpcUrl(superchainTarget)
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

	rollupConfig, err := constructChainConfig(rollupConfigPath, chainName, publicRPC, sequencerRPC, explorer, superchainLevel, standardChainCandidate, isFaultProofs)
	if err != nil {
		return fmt.Errorf("failed to construct rollup config: %w", err)
	}

	targetFilePath := filepath.Join(targetDir, chainShortName+".yaml")
	err = writeChainConfig(rollupConfig, targetFilePath)
	if err != nil {
		return fmt.Errorf("error generating chain config .yaml file: %w", err)
	}

	fmt.Printf("Wrote config for new chain with identifier %s", rollupConfig.Identifier())

	// Create genesis-system-config data
	// (this is deprecated, users should load this from L1, when available via SystemConfig)
	dirPath := filepath.Join(superchainRepoWritePath, "superchain", "extra", "genesis-system-configs", superchainTarget)

	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the genesis system config JSON to a new file
	systemConfigJSON, err := json.MarshalIndent(rollupConfig.Genesis.SystemConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal genesis system config json: %w", err)
	}

	filePath := filepath.Join(dirPath, chainShortName+".json")
	if err := os.WriteFile(filePath, systemConfigJSON, 0o644); err != nil {
		return fmt.Errorf("failed to write genesis system config json: %w", err)
	}
	fmt.Printf("Genesis system config written to: %s\n", filePath)

	err = readAddressesFromChain(addresses, l1RpcUrl, isFaultProofs)
	if err != nil {
		return fmt.Errorf("failed to read addresses from chain: %w", err)
	}

	if rollupConfig.Plasma != nil {
		addresses["DAChallengeAddress"] = rollupConfig.Plasma.DAChallengeAddress.String()
	}

	err = writeAddressesToJSON(addresses, superchainRepoWritePath, superchainTarget, chainShortName)
	if err != nil {
		return fmt.Errorf("failed to write contract addresses to JSON file: %w", err)
	}

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
