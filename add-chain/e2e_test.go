package main

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCLIApp(t *testing.T) {
	tests := []struct {
		name                   string
		chainName              string
		rollupConfigFile       string
		standardChainCandidate bool
		chainType              string
	}{
		{
			name:             "baseline",
			chainName:        "awesomechain_baseline",
			rollupConfigFile: "./testdata/monorepo/op-node/rollup_baseline.json",
			chainType:        "standard",
		},
		{
			name:             "plasma",
			chainName:        "awesomechain_plasma",
			rollupConfigFile: "./testdata/monorepo/op-node/rollup_plasma.json",
			chainType:        "standard",
		},
		{
			name:                   "standard-candidate",
			chainName:              "awesomechain_standard-candidate",
			rollupConfigFile:       "./testdata/monorepo/op-node/rollup_baseline.json",
			chainType:              "frontier",
			standardChainCandidate: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			args := []string{
				"add-chain",
				"--chain-type=" + tt.chainType,
				"--chain-name=" + tt.chainName,
				"--rollup-config=" + tt.rollupConfigFile,
				"--standard-chain-candidate=" + strconv.FormatBool(tt.standardChainCandidate),
				"--test=" + "true",
			}
			err := app.Run(args)
			require.NoError(t, err, "add-chain app failed")

			checkConfigYaml(t, tt.name, tt.chainName)
			compareJsonFiles(t, "./testdata/superchain/extra/addresses/sepolia/", tt.name, tt.chainName)
			compareJsonFiles(t, "./testdata/superchain/extra/genesis-system-configs/sepolia/", tt.name, tt.chainName)
		})
	}
}

func compareJsonFiles(t *testing.T, dirPath, testName, chainName string) {
	expectedBytes, err := os.ReadFile(dirPath + "expected_" + testName + ".json")
	require.NoError(t, err, "failed to read expected.json file from "+dirPath)

	var expectJSON map[string]interface{}
	err = json.Unmarshal(expectedBytes, &expectJSON)
	require.NoError(t, err, "failed to unmarshal expected.json file from "+dirPath)

	testBytes, err := os.ReadFile(dirPath + chainName + ".json")
	require.NoError(t, err, "failed to read test generated json file from "+dirPath)

	var testJSON map[string]interface{}
	err = json.Unmarshal(testBytes, &testJSON)
	require.NoError(t, err, "failed to read test generated json file from "+dirPath)

	require.Equal(t, expectJSON, testJSON, "test .json contents do not meet expectation")
}

func checkConfigYaml(t *testing.T, testName, chainName string) {
	expectedBytes, err := os.ReadFile("./testdata/superchain/configs/sepolia/expected_" + testName + ".yaml")
	require.NoError(t, err, "failed to read expected.yaml config file: %w", err)

	var expectedYaml map[string]interface{}
	err = yaml.Unmarshal(expectedBytes, &expectedYaml)
	require.NoError(t, err, "failed to unmarshal expected.yaml config file: %w", err)

	testBytes, err := os.ReadFile("./testdata/superchain/configs/sepolia/" + chainName + ".yaml")
	require.NoError(t, err, "failed to read awesomechain.yaml config file: %w", err)

	require.Equal(t, string(expectedBytes), string(testBytes), "test .yaml contents do not meet expectation")
}
