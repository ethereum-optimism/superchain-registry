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

func getSuperchainLevel(chainType string) (int, error) {
	switch chainType {
	case "standard":
		fmt.Printf("Adding standard chain to superchain-registry...\n\n")
		return 2, nil
	case "frontier":
		fmt.Printf("Adding frontier chain to superchain-registry...\n\n")
		return 1, nil
	}

	return 0, fmt.Errorf("invalid chain type: %s", chainType)
}
