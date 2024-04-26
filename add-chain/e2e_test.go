package main

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/urfave/cli/v2"
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
	if err != nil {
		t.Errorf("add-chain app failed: %v", err)
	}

	yamlEqual, err := checkConfigYaml()
	if err != nil {
		t.Errorf("failed to read yaml config files: %v", err)
	}
	if !yamlEqual {
		t.Error("test config yaml file does not match expected file")
	}

	jsonEqual, err := compareJsonFiles("./testdata/superchain/extra/addresses/sepolia/")
	if err != nil {
		t.Errorf("failed to read json address files: %v", err)
	}
	if !jsonEqual {
		t.Error("test json address file does not match expected file")
	}

	jsonEqual, err = compareJsonFiles("./testdata/superchain/extra/genesis-system-configs/sepolia/")
	if err != nil {
		t.Errorf("failed to read json genesis files: %v", err)
	}
	if !jsonEqual {
		t.Error("test json genesis file does not match expected file")
	}
}

func compareJsonFiles(dirPath string) (bool, error) {
	expectedBytes, err := os.ReadFile(dirPath + "expected.json")
	if err != nil {
		return false, err
	}

	var expectJSON map[string]interface{}
	if err := json.Unmarshal(expectedBytes, &expectJSON); err != nil {
		return false, err
	}

	testBytes, err := os.ReadFile(dirPath + "awesomechain.json")
	if err != nil {
		return false, err
	}

	var testJSON map[string]interface{}
	if err := json.Unmarshal(testBytes, &testJSON); err != nil {
		return false, err
	}

	return reflect.DeepEqual(expectJSON, testJSON), nil
}

func checkConfigYaml() (bool, error) {
	expectedBytes, err := os.ReadFile("./testdata/superchain/configs/sepolia/expected.yaml")
	if err != nil {
		return false, err
	}

	testBytes, err := os.ReadFile("./testdata/superchain/configs/sepolia/awesomechain.yaml")
	if err != nil {
		return false, err
	}

	return bytes.Equal(expectedBytes, testBytes), nil
}
