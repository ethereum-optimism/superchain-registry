package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func TestCLIApp(t *testing.T) {
	app := &cli.App{
		Name:   "add-chain",
		Usage:  "Add a new chain to the superchain-registry",
		Flags:  []cli.Flag{ChainTypeFlag, ChainNameFlag, RollupConfigFlag, TestFlag},
		Action: entrypoint,
	}

	tests := []struct {
		name             string
		chainName        string
		rollupConfigFile string
	}{
		{
			name:             "baseline",
			chainName:        "awesomechain_baseline",
			rollupConfigFile: "./testdata/monorepo/op-node/rollup_baseline.json",
		},
		{
			name:             "plasma",
			chainName:        "awesomechain_plasma",
			rollupConfigFile: "./testdata/monorepo/op-node/rollup_plasma.json",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			args := []string{"add-chain", "-chain-type", "standard", "-chain-name", tt.chainName, "-rollup-config", tt.rollupConfigFile, "-test", "true"}
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

	require.True(t, reflect.DeepEqual(expectJSON, testJSON), "test .json contents do not meet expectation:\n %s", string(testBytes))
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
