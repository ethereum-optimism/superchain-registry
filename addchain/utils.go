package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func executeCommand(command string, args []string) (string, error) {
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

func inferRpcUrl(superchainTarget string) (string, error) {
	switch superchainTarget {
	case "mainnet":
		return "https://ethereum-mainnet-rpc.allthatnode.com", nil
	case "sepolia":
		return "https://ethereum-sepolia-rpc.allthatnode.com", nil
	}
	return "", fmt.Errorf("Unsupported Superchain Target %s\n", superchainTarget)
}

func getSuperchainLevel(chainType string) (int, error) {
	switch chainType {
	case "standard":
		fmt.Println("Adding standard chain to superchain-registry...")
		return 2, nil
	case "frontier":
		fmt.Println("Adding frontier chain to superchain-registry...")
		return 1, nil
	}

	return 0, fmt.Errorf("Invalid chain type: %s", chainType)
}
