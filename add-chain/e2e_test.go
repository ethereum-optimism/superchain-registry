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

var tests = []struct {
	name                   string
	chainName              string
	rollupConfigFile       string
	standardChainCandidate bool
	chainType              string
	deploymentsDir         string
}{
	{
		name:             "baseline",
		chainName:        "testchain_baseline",
		rollupConfigFile: "./testdata/monorepo/op-node/rollup_baseline.json",
		chainType:        "standard",
	},
	{
		name:             "plasma",
		chainName:        "testchain_plasma",
		rollupConfigFile: "./testdata/monorepo/op-node/rollup_plasma.json",
		chainType:        "standard",
	},
	{
		name:                   "standard-candidate",
		chainName:              "testchain_standard-candidate",
		rollupConfigFile:       "./testdata/monorepo/op-node/rollup_standard-candidate.json",
		chainType:              "frontier",
		standardChainCandidate: true,
	},
	{
		name:             "faultproofs",
		chainName:        "testchain_faultproofs",
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

			cleanupTestFiles(t, tt.chainName)

			args := []string{
				"add-chain",
				"--chain-type=" + tt.chainType,
				"--chain-name=" + tt.chainName,
				"--rollup-config=" + tt.rollupConfigFile,
				"--standard-chain-candidate=" + strconv.FormatBool(tt.standardChainCandidate),
				"--test=true",
			}

			if tt.deploymentsDir != "" {
				args = append(args, "--deployments-dir="+tt.deploymentsDir)
			}

			err := app.Run(args)
			require.NoError(t, err, "add-chain app failed")

			checkConfigYaml(t, tt.name, tt.chainName)
			compareJsonFiles(t, "superchain/extra/addresses/sepolia/", tt.name, tt.chainName)
			compareJsonFiles(t, "superchain/extra/genesis-system-configs/sepolia/", tt.name, tt.chainName)
		})
	}
}

func TestAddChain_CheckRollupConfig(t *testing.T) {
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Must run the following subcommand after the main command completes so that when the op-node/rollup
			// package imports the superchain package, the superchain package includes the output test files
			args := []string{
				"add-chain",
				"check-rollup-config",
				"--rollup-config=" + tt.rollupConfigFile,
			}
			err := app.Run(args)
			require.NoError(t, err, "add-chain check-rollup-config failed")
		})
	}
}

func compareJsonFiles(t *testing.T, dirPath, testName, chainName string) {
	expectedBytes, err := os.ReadFile("./testdata/" + dirPath + "expected_" + testName + ".json")
	require.NoError(t, err, "failed to read expected.json file from "+dirPath)

	var expectJSON map[string]interface{}
	err = json.Unmarshal(expectedBytes, &expectJSON)
	require.NoError(t, err, "failed to unmarshal expected.json file from "+dirPath)

	testBytes, err := os.ReadFile("../" + dirPath + chainName + ".json")
	require.NoError(t, err, "failed to read test generated json file from "+dirPath)

	var testJSON map[string]interface{}
	err = json.Unmarshal(testBytes, &testJSON)
	require.NoError(t, err, "failed to read test generated json file from "+dirPath)

	diff := cmp.Diff(expectJSON, testJSON)
	require.Equal(t, diff, "", "expected json (-) does not match test json (+): %s", diff)
}

func checkConfigYaml(t *testing.T, testName, chainName string) {
	expectedBytes, err := os.ReadFile("./testdata/superchain/configs/sepolia/expected_" + testName + ".yaml")
	require.NoError(t, err, "failed to read expected.yaml config file: %w", err)

	var expectedYaml map[string]interface{}
	err = yaml.Unmarshal(expectedBytes, &expectedYaml)
	require.NoError(t, err, "failed to unmarshal expected.yaml config file: %w", err)

	testBytes, err := os.ReadFile("../superchain/configs/sepolia/" + chainName + ".yaml")
	require.NoError(t, err, "failed to read testchain.yaml config file: %w", err)

	require.Equal(t, string(expectedBytes), string(testBytes), "test .yaml contents do not meet expectation")
}

func cleanupTestFiles(t *testing.T, chainName string) {
	paths := []string{
		"../superchain/extra/addresses/sepolia/" + chainName + ".json",
		"../superchain/extra/genesis-system-configs/sepolia/" + chainName + ".json",
		"../superchain/configs/sepolia/" + chainName + ".yaml",
	}

	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			// Log the error if it's something other than "file does not exist"
			t.Logf("Error removing file %s: %v\n", path, err)
		}
	}
	t.Logf("Removed test artifacts for chain: %s", chainName)
}
