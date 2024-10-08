package genesis

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/common"
)

func executeCommandInDir(dir string, cmd *exec.Cmd) error {
	log.Printf("executing %s", cmd.String())
	cmd.Dir = dir
	var outErr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &outErr
	err := cmd.Run()
	if err != nil {
		// error case : status code of command is different from 0
		fmt.Println(outErr.String())
	}
	return err
}

func mustExecuteCommandInDir(dir string, cmd *exec.Cmd) {
	err := executeCommandInDir(dir, cmd)
	if err != nil {
		panic(err)
	}
}

func streamOutputToLogger(reader io.Reader, t *testing.T) {
	scanner := bufio.NewScanner(reader)

	// default scan buffer in bufio is 64k, which limits the max token size that
	// can be scanned. Since we are handling allocs which contain large hex strings,
	// the output of a diff must handle large tokens.
	// Provide our own buffer per recommendation in bufio before calling Scan()
	// ref: https://github.com/golang/go/blob/release-branch.go1.22/src/bufio/scan.go#L78-L82
	buf := [bufio.MaxScanTokenSize * 1000]byte{}
	scanner.Buffer(buf[:], bufio.MaxScanTokenSize*1000)

	for scanner.Scan() {
		t.Log(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		t.Errorf("Error reading command output: %v", err)
	}
}

func writeDeployments(chainId uint64, directory string) error {
	as := Addresses[chainId]

	data, err := json.Marshal(as)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(directory, ".deploy"), data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func writeDeploymentsLegacy(chainId uint64, directory string) error {
	// Prepare a HardHat Deployment type, we need this whole structure to make things
	// work, although it is only the Address field which ends up getting used.
	type StorageLayoutEntry struct {
		AstId    uint   `json:"astId"`
		Contract string `json:"contract"`
		Label    string `json:"label"`
		Offset   uint   `json:"offset"`
		Slot     uint   `json:"slot,string"`
		Type     string `json:"type"`
	}
	type StorageLayoutType struct {
		Encoding      string `json:"encoding"`
		Label         string `json:"label"`
		NumberOfBytes uint   `json:"numberOfBytes,string"`
		Key           string `json:"key,omitempty"`
		Value         string `json:"value,omitempty"`
		Base          string `json:"base,omitempty"`
	}
	type StorageLayout struct {
		Storage []StorageLayoutEntry         `json:"storage"`
		Types   map[string]StorageLayoutType `json:"types"`
	}
	type Deployment struct {
		Name             string
		Abi              []string        `json:"abi"`
		Address          string          `json:"address"`
		Args             []any           `json:"args"`
		Bytecode         string          `json:"bytecode"`
		DeployedBytecode string          `json:"deployedBytecode"`
		Devdoc           json.RawMessage `json:"devdoc"`
		Metadata         string          `json:"metadata"`
		Receipt          json.RawMessage `json:"receipt"`
		SolcInputHash    string          `json:"solcInputHash"`
		StorageLayout    StorageLayout   `json:"storageLayout"`
		TransactionHash  common.Hash     `json:"transactionHash"`
		Userdoc          json.RawMessage `json:"userdoc"`
	}

	// Initialize your struct with some data
	data := Addresses[chainId]

	type AddressList2 AddressList // use another type to prevent infinite recursion later on
	b := AddressList2(*data)

	o, err := json.Marshal(b)
	if err != nil {
		return err
	}

	out := make(map[string]Address)
	err = json.Unmarshal(o, &out)
	if err != nil {
		return err
	}

	for k, v := range out {
		text, err := v.MarshalText()
		if err != nil || !strings.HasPrefix(string(text), "0x") {
			continue
		}
		// Define the Deployment object, filling in only what we need
		jsonData := Deployment{Address: v.String(), Name: k}

		raw, err := json.MarshalIndent(jsonData, "", " ")
		if err != nil {
			return err
		}

		fileName := fmt.Sprintf("%s.json", k)
		file, err := os.Create(path.Join(directory, fileName))
		if err != nil {
			return fmt.Errorf("failed to create file for field %s: %w", k, err)
		}
		defer file.Close()

		// Write the JSON content to the file
		_, err = file.Write(raw)
		if err != nil {
			return fmt.Errorf("failed to write JSON to file for field %s: %w", k, err)
		}

		fmt.Printf("Created file: %s\n", fileName)
	}
	return nil
}
