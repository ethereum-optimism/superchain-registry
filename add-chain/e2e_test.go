package main

import (
	"bytes"
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
		Flags:  []cli.Flag{ChainTypeFlag, TestFlag},
		Action: entrypoint,
	}

	args := []string{"add-chain", "-chain-type", "standard", "-test", "true"}
	err := app.Run(args)
	require.NoError(t, err, "add-chain app failed")

	checkConfigYaml(t)
	compareJsonFiles(t, "./testdata/superchain/extra/addresses/sepolia/")
	compareJsonFiles(t, "./testdata/superchain/extra/genesis-system-configs/sepolia/")
}

func compareJsonFiles(t *testing.T, dirPath string) {
	expectedBytes, err := os.ReadFile(dirPath + "expected.json")
	require.NoError(t, err, "failed to read expected.json file from "+dirPath)

	var expectJSON map[string]interface{}
	err = json.Unmarshal(expectedBytes, &expectJSON)
	require.NoError(t, err, "failed to unmarshal expected.json file from "+dirPath)

	testBytes, err := os.ReadFile(dirPath + "awesomechain.json")
	require.NoError(t, err, "failed to read awesomechain.json file from "+dirPath)

	var testJSON map[string]interface{}
	err = json.Unmarshal(testBytes, &testJSON)
	require.NoError(t, err, "failed to read awesomechain.json file from "+dirPath)

	require.True(t, reflect.DeepEqual(expectJSON, testJSON), "awesomechain.json contents do not meet expectation:\n %s", string(testBytes))
}

func checkConfigYaml(t *testing.T) {
	expectedBytes, err := os.ReadFile("./testdata/superchain/configs/sepolia/expected.yaml")
	require.NoError(t, err, "failed to read expected.yaml config file: %w", err)

	var expectedYaml map[string]interface{}
	err = yaml.Unmarshal(expectedBytes, &expectedYaml)
	require.NoError(t, err, "failed to unmarshal expected.yaml config file: %w", err)

	testBytes, err := os.ReadFile("./testdata/superchain/configs/sepolia/awesomechain.yaml")
	require.NoError(t, err, "failed to read awesomechain.yaml config file: %w", err)

	require.True(t, bytes.Equal(expectedBytes, testBytes), "awesomechain.yaml contents do not meet expectation:\n %s", string(testBytes))
}
