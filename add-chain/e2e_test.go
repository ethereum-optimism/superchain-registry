package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

const (
	addressDir          = "../superchain/extra/addresses/sepolia/"
	configDir           = "../superchain/configs/sepolia/"
	validationInputsDir = "../validation/genesis/validation-inputs"
)

var tests = []struct {
	name                   string
	chainName              string
	chainId                uint64
	chainShortName         string
	rollupConfigFile       string
	standardChainCandidate bool
	chainType              string
	deploymentsDir         string
}{
	{
		name:             "baseline",
		chainName:        "testchain_baseline",
		chainId:          4206900,
		chainShortName:   "testchain_b",
		rollupConfigFile: "./testdata/monorepo/op-node/rollup_baseline.json",
		deploymentsDir:   "./testdata/monorepo/deployments",
	},
	{
		name:             "baseline_legacy",
		chainName:        "testchain_baseline_legacy",
		chainId:          4206905,
		chainShortName:   "testchain_bl",
		rollupConfigFile: "./testdata/monorepo/op-node/rollup_baseline_legacy.json",
		deploymentsDir:   "./testdata/monorepo/deployments-legacy",
	},
	{
		name:                   "zorasep",
		chainName:              "testchain_zorasep",
		chainId:                4206904,
		chainShortName:         "testchain_zs",
		rollupConfigFile:       "./testdata/monorepo/op-node/rollup_zorasep.json",
		deploymentsDir:         "./testdata/monorepo/deployments",
		standardChainCandidate: true,
	},
	{
		name:                   "altda",
		chainName:              "testchain_altda",
		chainId:                4206901,
		chainShortName:         "testchain_ad",
		rollupConfigFile:       "./testdata/monorepo/op-node/rollup_altda.json",
		deploymentsDir:         "./testdata/monorepo/deployments",
		standardChainCandidate: true,
	},
	{
		name:                   "standard-candidate",
		chainName:              "testchain_standard-candidate",
		chainId:                4206902,
		chainShortName:         "testchain_sc",
		rollupConfigFile:       "./testdata/monorepo/op-node/rollup_standard-candidate.json",
		deploymentsDir:         "./testdata/monorepo/deployments",
		standardChainCandidate: true,
	},
	{
		name:                   "faultproofs",
		chainName:              "testchain_faultproofs",
		chainId:                4206903,
		chainShortName:         "testchain_fp",
		rollupConfigFile:       "./testdata/monorepo/op-node/rollup_faultproofs.json",
		deploymentsDir:         "./testdata/monorepo/deployments-faultproofs",
		standardChainCandidate: true,
	},
}

func TestAddChain_Main(t *testing.T) {
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cleanupTestFiles(t, tt.chainShortName, tt.chainId)

			err := os.Setenv("SCR_RUN_TESTS", "true")
			require.NoError(t, err, "failed to set SCR_RUN_TESTS env var")

			args := []string{
				"add-chain",
				"--chain-name=" + tt.chainName,
				"--chain-short-name=" + tt.chainShortName,
				"--rollup-config=" + tt.rollupConfigFile,
				"--genesis=" + "./testdata/monorepo/op-node/genesis_zorasep.json",
				"--deployments-dir=" + tt.deploymentsDir,
				"--standard-chain-candidate=" + strconv.FormatBool(tt.standardChainCandidate),
			}

			err = runApp(args)
			require.NoError(t, err, "add-chain app failed")

			checkConfigTOML(t, tt.name, tt.chainShortName)
		})
	}

	t.Run("compress-genesis", func(t *testing.T) {
		// Must run this test to produce the .json.gz output artifact for the
		// subsequent TestAddChain_CheckGenesis
		t.Parallel()
		err := os.Setenv("SCR_RUN_TESTS", "true")
		require.NoError(t, err, "failed to set SCR_RUN_TESTS env var")

		args := []string{
			"add-chain",
			"compress-genesis",
			"--genesis=" + "./testdata/monorepo/op-node/genesis_zorasep.json",
			"--superchain-target=" + "sepolia",
			"--chain-short-name=" + "testchain_zs",
		}
		err = runApp(args)
		require.NoError(t, err, "add-chain compress-genesis failed")
	})
}

func TestAddChain_CheckRollupConfig(t *testing.T) {
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Must run the following subcommand after the main command completes so that when the op-node/rollup
			// package imports the superchain package, the superchain package includes the output test files
			err := os.Setenv("SCR_RUN_TESTS", "true")
			require.NoError(t, err, "failed to set SCR_RUN_TESTS env var")

			args := []string{
				"add-chain",
				"check-rollup-config",
				"--rollup-config=" + tt.rollupConfigFile,
			}
			err = runApp(args)
			require.NoError(t, err, "add-chain check-rollup-config failed")
		})
	}
}

func TestAddChain_CheckGenesis(t *testing.T) {
	t.Run("genesis_zorasep", func(t *testing.T) {
		err := os.Setenv("SCR_RUN_TESTS", "true")
		require.NoError(t, err, "failed to set SCR_RUN_TESTS env var")

		args := []string{
			"add-chain",
			"check-genesis",
			"--genesis=" + "./testdata/monorepo/op-node/genesis_zorasep.json",
		}
		err = runApp(args)
		require.NoError(t, err, "add-chain check-genesis failed")
	})
}

func checkConfigTOML(t *testing.T, testName, chainShortName string) {
	expectedBytes, err := os.ReadFile("./testdata/superchain/configs/sepolia/expected_" + testName + ".toml")
	require.NoError(t, err, "failed to read expected.toml config file: %w", err)

	var expectedTOML map[string]interface{}
	err = toml.Unmarshal(expectedBytes, &expectedTOML)
	require.NoError(t, err, "failed to unmarshal expected.toml config file: %w", err)

	testBytes, err := os.ReadFile(configDir + chainShortName + ".toml")
	require.NoError(t, err, "failed to read testchain.toml config file: %w", err)

	require.Equal(t, string(expectedBytes), string(testBytes), "test .toml contents do not meet expectation")
}

func cleanupTestFiles(t *testing.T, chainShortName string, chainId uint64) {
	paths := []string{
		addressDir + chainShortName + ".json",
		configDir + chainShortName + ".toml",
	}

	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			// Log the error if it's something other than "file does not exist"
			t.Logf("Error removing file %s: %v\n", path, err)
		}
	}

	folderName := fmt.Sprintf("%d", chainId)
	folderName = folderName + "-test"
	genesisValidationInputs := filepath.Join(validationInputsDir, folderName)
	err := os.RemoveAll(genesisValidationInputs)
	if err != nil {
		t.Logf("Error removing genesis validation inputs %s: %v\n", genesisValidationInputs, err)
	}
	t.Logf("Removed test artifacts for chain: %s", chainShortName)
}
