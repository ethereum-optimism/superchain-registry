package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/optimism/op-e2e/bindings"
	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum-optimism/superchain-registry/ops/config"
	"github.com/ethereum-optimism/superchain-registry/ops/flags"
	"github.com/ethereum-optimism/superchain-registry/ops/utils"
	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum-optimism/superchain-registry/validation/genesis"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
)

var AddNewChainCmd = cli.Command{
	Name:  "add-new-chain",
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
	Action: func(c *cli.Context) error {
		standardChainCandidate := c.Bool(flags.StandardChainCandidateFlag.Name)

		superchainLevel := superchain.Frontier // All chains enter as frontier chains

		publicRPC := c.String(flags.PublicRpcFlag.Name)
		sequencerRPC := c.String(flags.SequencerRpcFlag.Name)
		explorer := c.String(flags.ExplorerFlag.Name)
		superchainTarget := c.String(flags.SuperchainTargetFlag.Name)
		monorepoDir := c.String(flags.MonorepoDirFlag.Name)

		chainName := c.String(flags.ChainNameFlag.Name)
		rollupConfigPath := c.String(flags.RollupConfigFlag.Name)
		genesisPath := c.String(flags.GenesisFlag.Name)
		deployConfigPath := c.String(flags.DeployConfigFlag.Name)
		genesisCreationCommit := c.String(flags.GenesisCreationCommit.Name)
		deploymentsDir := c.String(flags.DeploymentsDirFlag.Name)
		chainShortName := c.String(flags.ChainShortNameFlag.Name)

		// Get the current script filepath
		_, thisFile, _, ok := runtime.Caller(0)
		if !ok {
			panic("error getting current filepath")
		}
		superchainRepoRoot := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))

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
		err = utils.ReadAddressesFromJSON(&addresses, deploymentsDir)
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

		fmt.Printf("✅ Rollup config successfully constructed\n")

		err = utils.ReadAddressesFromChain(&addresses, l1RpcUrl, isFaultProofs)
		if err != nil {
			return fmt.Errorf("failed to read addresses from chain: %w", err)
		}

		fmt.Printf("✅ Addresses read from chain\n")

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
			return fmt.Errorf("error generating chain config %s.toml file: %w", chainShortName, err)
		}

		fmt.Printf("✅ Wrote config for new chain to %s\n", targetFilePath)

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
		fmt.Printf("✅ Copied deploy-config json file to validation module\n")

		err = writeGenesisValidationMetadata(genesisCreationCommit, genesisValidationInputsDir)
		if err != nil {
			return fmt.Errorf("error writing genesis validation metadata file: %w", err)
		}
		fmt.Printf("✅ Wrote genesis validation metadata file\n")

		return nil
	},
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
	castResult, err := validation.CastCall(optimismPortalProxyAddress, "version()(string)", nil, l1RpcUrl)
	if err != nil {
		return false, fmt.Errorf("failed to get OptimismPortalProxy.version(): %w", err)
	}

	version, err := strconv.Unquote(castResult[0])
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

	usingCustomGasToken, err := retry.Do(context.Background(), 3, retry.Exponential(), func() (bool, error) {
		usingCustomGasToken, err := sc.IsCustomGasToken(&opts)
		if err != nil {
			if strings.Contains(err.Error(), "execution reverted") {
				// This happens when the SystemConfig contract
				// does not yet have the CGT functionality.
				return false, nil
			}
			return false, err
		}
		return usingCustomGasToken, nil
	})
	if err != nil {
		return nil, err
	}
	if !usingCustomGasToken {
		return nil, nil
	}

	result, err := retry.Do(context.Background(), 3, retry.Exponential(), func() (struct {
		Addr     common.Address
		Decimals uint8
	}, error,
	) {
		var zeroVal struct {
			Addr     common.Address
			Decimals uint8
		}
		result, err := sc.GasPayingToken(&opts)
		if err != nil {
			if strings.Contains(err.Error(), "execution reverted") {
				// This happens when the SystemConfig contract
				// does not yet have the CGT functionality.
				return zeroVal, nil
			}
			return zeroVal, err
		}
		return result, nil
	})
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
	// involving OPLabs engineers after the add-new-chain command runs.
	const defaultNodeVersion = "18.12.1"
	const defaultMonorepoBuildCommand = "pnpm"
	const defaultGenesisCreationCommand = "forge1" // See validation/genesis/commands.go
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
