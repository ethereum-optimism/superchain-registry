package main

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const (
	addressDir       = "../superchain/extra/addresses/sepolia/"
	ymlConfigDir     = "../superchain/configs/sepolia/"
	genesisConfigDir = "../superchain/extra/genesis-system-configs/sepolia/"
)

var tests = []struct {
	name                   string
	chainName              string
	chainShortName         string
	rollupConfigFile       string
	standardChainCandidate bool
	chainType              string
	deploymentsDir         string
}{
	{
		name:             "baseline",
		chainName:        "testchain_baseline",
		chainShortName:   "testchain_b",
		rollupConfigFile: "./testdata/monorepo/op-node/rollup_baseline.json",
		deploymentsDir:   "./testdata/monorepo/deployments",
		chainType:        "standard",
	},
	{
		name:                   "zorasep",
		chainName:              "testchain_zorasep",
		chainShortName:         "testchain_zs",
		rollupConfigFile:       "./testdata/monorepo/op-node/rollup_zorasep.json",
		deploymentsDir:         "./testdata/monorepo/deployments",
		chainType:              "frontier",
		standardChainCandidate: true,
	},
	{
		name:             "plasma",
		chainName:        "testchain_plasma",
		chainShortName:   "testchain_p",
		rollupConfigFile: "./testdata/monorepo/op-node/rollup_plasma.json",
		deploymentsDir:   "./testdata/monorepo/deployments",
		chainType:        "standard",
	},
	{
		name:                   "standard-candidate",
		chainName:              "testchain_standard-candidate",
		chainShortName:         "testchain_sc",
		rollupConfigFile:       "./testdata/monorepo/op-node/rollup_standard-candidate.json",
		chainType:              "frontier",
		deploymentsDir:         "./testdata/monorepo/deployments",
		standardChainCandidate: true,
	},
	{
		name:             "faultproofs",
		chainName:        "testchain_faultproofs",
		chainShortName:   "testchain_fp",
		rollupConfigFile: "./testdata/monorepo/op-node/rollup_faultproofs.json",
		chainType:        "standard",
		deploymentsDir:   "./testdata/monorepo/deployments-faultproofs",
	},
}

func TestAddChain_Main(t *testing.T) {
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cleanupTestFiles(t, tt.chainShortName)

			err := os.Setenv("SCR_RUN_TESTS", "true")
			require.NoError(t, err, "failed to set SCR_RUN_TESTS env var")

			args := []string{
				"add-chain",
				"--chain-type=" + tt.chainType,
				"--chain-name=" + tt.chainName,
				"--chain-short-name=" + tt.chainShortName,
				"--rollup-config=" + tt.rollupConfigFile,
				"--deployments-dir=" + tt.deploymentsDir,
				"--standard-chain-candidate=" + strconv.FormatBool(tt.standardChainCandidate),
			}

			err = runApp(args)
			require.NoError(t, err, "add-chain app failed")

			checkConfigYaml(t, tt.name, tt.chainShortName)
			compareJsonFiles(t, "superchain/extra/addresses/sepolia/", tt.name, tt.chainShortName)
			compareJsonFiles(t, "superchain/extra/genesis-system-configs/sepolia/", tt.name, tt.chainShortName)
		})
	}

	t.Run("compress-genesis", func(t *testing.T) {
		// Must run this test to produce the .json.gz output artifact for the
		// subsequent CheckGenesisConfig test
		t.Parallel()
		err := os.Setenv("SCR_RUN_TESTS", "true")
		require.NoError(t, err, "failed to set SCR_RUN_TESTS env var")

		args := []string{
			"add-chain",
			"compress-genesis",
			"--l2-genesis=" + "./testdata/monorepo/op-node/genesis_zorasep.json",
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

func TestAddChain_CheckGenesisConfig(t *testing.T) {
	t.Run("genesis_zorasep", func(t *testing.T) {
		t.Parallel()
		err := os.Setenv("SCR_RUN_TESTS", "true")
		require.NoError(t, err, "failed to set SCR_RUN_TESTS env var")

		args := []string{
			"add-chain",
			"check-genesis-config",
			"--genesis-config=" + "./testdata/monorepo/op-node/genesis_zorasep.json",
			"--chain-id=" + "4206904",
		}
		err = runApp(args)
		require.NoError(t, err, "add-chain check-genesis-config failed")
	})
}

func compareJsonFiles(t *testing.T, dirPath, testName, chainShortName string) {
	expectedBytes, err := os.ReadFile("./testdata/" + dirPath + "expected_" + testName + ".json")
	require.NoError(t, err, "failed to read expected.json file from "+dirPath)

	var expectJSON map[string]interface{}
	err = json.Unmarshal(expectedBytes, &expectJSON)
	require.NoError(t, err, "failed to unmarshal expected.json file from "+dirPath)

	testBytes, err := os.ReadFile("../" + dirPath + chainShortName + ".json")
	require.NoError(t, err, "failed to read test generated json file from "+dirPath)

	var testJSON map[string]interface{}
	err = json.Unmarshal(testBytes, &testJSON)
	require.NoError(t, err, "failed to read test generated json file from "+dirPath)

	diff := cmp.Diff(expectJSON, testJSON)
	require.Equal(t, diff, "", "expected json (-) does not match test json (+): %s", diff)
}

func checkConfigYaml(t *testing.T, testName, chainShortName string) {
	expectedBytes, err := os.ReadFile("./testdata/superchain/configs/sepolia/expected_" + testName + ".yaml")
	require.NoError(t, err, "failed to read expected.yaml config file: %w", err)

	var expectedYaml map[string]interface{}
	err = yaml.Unmarshal(expectedBytes, &expectedYaml)
	require.NoError(t, err, "failed to unmarshal expected.yaml config file: %w", err)

	testBytes, err := os.ReadFile(ymlConfigDir + chainShortName + ".yaml")
	require.NoError(t, err, "failed to read testchain.yaml config file: %w", err)

	require.Equal(t, string(expectedBytes), string(testBytes), "test .yaml contents do not meet expectation")
}

func cleanupTestFiles(t *testing.T, chainShortName string) {
	paths := []string{
		addressDir + chainShortName + ".json",
		genesisConfigDir + chainShortName + ".json",
		ymlConfigDir + chainShortName + ".yaml",
	}

	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			// Log the error if it's something other than "file does not exist"
			t.Logf("Error removing file %s: %v\n", path, err)
		}
	}
	t.Logf("Removed test artifacts for chain: %s", chainShortName)
}
