package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func createExtraGenesisData(superchainRepoPath, superchainTarget, monorepoDir, genesisConfig, chainName string) error {
	// Create directory for genesis data
	genesisDir := filepath.Join(superchainRepoPath, "superchain", "extra", "genesis", superchainTarget)
	if err := os.MkdirAll(genesisDir, 0755); err != nil {
		return fmt.Errorf("failed to create genesis directory: %w", err)
	}

	// Set current working directory to the monorepo directory
	if err := os.Chdir(monorepoDir); err != nil {
		return fmt.Errorf("failed to change directory to monorepo: %w", err)
	}

	cmd := exec.Command("go", "run", "./op-chain-ops/cmd/registry-data",
		"--l2-genesis", genesisConfig,
		"--bytecodes-dir", filepath.Join(superchainRepoPath, "superchain", "extra", "bytecodes"),
		"--output", filepath.Join(genesisDir, chainName+".json.gz"))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run registry-data command: %w", err)
	}
	fmt.Println("Ran go program: <monorepo>/op-chain-ops/cmd/registry-data")

	return nil
}
