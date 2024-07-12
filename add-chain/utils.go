package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func castCall(contractAddress, calldata, l1RpcUrl string) (string, error) {
	cmd := exec.Command("cast", "call", contractAddress, calldata, "-r", l1RpcUrl)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s, %w", &stderr, err)
	}

	address := strings.Join(strings.Fields(out.String()), "") // remove whitespace
	if address == "" || address == "0x" {
		return "", fmt.Errorf("cast call returned empty address")
	}

	return address, nil
}
