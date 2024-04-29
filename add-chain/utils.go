package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ethereum-optimism/superchain-registry/superchain"
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

func getSuperchainLevel(chainType string) (superchain.SuperchainLevel, error) {
	switch chainType {
	case "standard":
		fmt.Printf("Adding standard chain to superchain-registry...\n\n")
		return superchain.Standard, nil
	case "frontier":
		fmt.Printf("Adding frontier chain to superchain-registry...\n\n")
		return superchain.Frontier, nil
	}

	return 0, fmt.Errorf("invalid chain type: %s", chainType)
}
