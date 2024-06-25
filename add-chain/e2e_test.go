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
		chainShortName         string
		rollupConfigFile       string
		standardChainCandidate bool
		chainType              string
		deploymentsDir         string
	}{
		{
			name:             "baseline",
			chainName:        "awesomechain_baseline",
			chainShortName:   "awsm_baseline",
			rollupConfigFile: "./testdata/monorepo/op-node/rollup_baseline.json",
			chainType:        "standard",
		},
		{
			name:             "plasma",
			chainName:        "awesomechain_plasma",
			chainShortName:   "awsm_plasma",
			rollupConfigFile: "./testdata/monorepo/op-node/rollup_plasma.json",
			chainType:        "standard",
		},
		{
			name:                   "standard-candidate",
			chainName:              "awesomechain_standard-candidate",
			chainShortName:         "awsm_standard-candidate",
			rollupConfigFile:       "./testdata/monorepo/op-node/rollup_baseline.json",
			chainType:              "frontier",
			standardChainCandidate: true,
		},
		{
			name:             "faultproofs",
			chainName:        "awesomechain_faultproofs",
			chainShortName:   "awsm_faultproofs",
			rollupConfigFile: "./testdata/monorepo/op-node/rollup_faultproofs.json",
			chainType:        "standard",
			deploymentsDir:   "./testdata/monorepo/deployments-faultproofs",
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
				"--chain-short-name=" + tt.chainShortName,
				"--rollup-config=" + tt.rollupConfigFile,
				"--standard-chain-candidate=" + strconv.FormatBool(tt.standardChainCandidate),
				"--test=" + "true",
				"--test=true",
			}

			if tt.deploymentsDir != "" {
				args = append(args, "--deployments-dir="+tt.deploymentsDir)
			}

			err := app.Run(args)
			require.NoError(t, err, "add-chain app failed")

			checkConfigYaml(t, tt.name, tt.chainShortName)
			compareJsonFiles(t, "./testdata/superchain/extra/addresses/sepolia/", tt.name, tt.chainShortName)
			compareJsonFiles(t, "./testdata/superchain/extra/genesis-system-configs/sepolia/", tt.name, tt.chainShortName)
		})
	}
}

func compareJsonFiles(t *testing.T, dirPath, testName, chainShortName string) {
	expectedBytes, err := os.ReadFile(dirPath + "expected_" + testName + ".json")
	require.NoError(t, err, "failed to read expected.json file from "+dirPath)

	var expectJSON map[string]interface{}
	err = json.Unmarshal(expectedBytes, &expectJSON)
	require.NoError(t, err, "failed to unmarshal expected.json file from "+dirPath)

	testBytes, err := os.ReadFile(dirPath + chainShortName + ".json")
	require.NoError(t, err, "failed to read test generated json file from "+dirPath)

	var testJSON map[string]interface{}
	err = json.Unmarshal(testBytes, &testJSON)
	require.NoError(t, err, "failed to read test generated json file from "+dirPath)

	require.Equal(t, expectJSON, testJSON, "test .json contents do not meet expectation")
}

func checkConfigYaml(t *testing.T, testName, chainShortName string) {
	expectedBytes, err := os.ReadFile("./testdata/superchain/configs/sepolia/expected_" + testName + ".yaml")
	require.NoError(t, err, "failed to read expected.yaml config file: %w", err)

	var expectedYaml map[string]interface{}
	err = yaml.Unmarshal(expectedBytes, &expectedYaml)
	require.NoError(t, err, "failed to unmarshal expected.yaml config file: %w", err)

	testBytes, err := os.ReadFile("./testdata/superchain/configs/sepolia/" + chainShortName + ".yaml")
	require.NoError(t, err, "failed to read config file: %s", err)

	require.Equal(t, string(expectedBytes), string(testBytes), "test .yaml contents do not meet expectation")
}
