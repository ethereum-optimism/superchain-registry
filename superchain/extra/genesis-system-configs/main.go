package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
)

// Function reads the chainlist to get the chain ids for the given network-name
// genesis system configs are generated for the given chain ids
func main() {
	// Read the chainlist
	currentDir := currentDir()
	repositoryRoot := filepath.Join(currentDir, "../../../")
	chainlistPath := filepath.Join(repositoryRoot, "chainList.json")
	chainlist, err := os.ReadFile(chainlistPath)
	if err != nil {
		fmt.Printf("Error reading chainlist file %s. Please generate chainList.json before executing this script.\n", chainlistPath)
		panic(err)
	}

	var chains []ChainEntry
	err = json.Unmarshal(chainlist, &chains)
	if err != nil {
		panic(err)
	}

	// For each sub directory in this directory, generate genesis system configs for each chain id file.
	dirs, err := os.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading directory")
		panic(err)
	}

	// Map chain id to genesis system config
	gscMap := make(map[int]GenesisSystemConfig)
	for _, dir := range dirs {
		if dir.IsDir() {
			parent := dir.Name()
			// Read the genesis system config files in the directory
			files, err := os.ReadDir(parent)
			if err != nil {
				fmt.Println("Error reading subdirectory:", parent)
				panic(err)
			}

			generated := 0
			for _, file := range files {
				name := file.Name()
				if filepath.Ext(name) == ".json" {
					name = name[:len(name)-5]

					// Find the chain id for the given parent and name
					preGen := generated
					for _, chain := range chains {
						identifier := fmt.Sprintf("%s/%s", parent, name)
						if chain.Identifier == identifier {
							// Read the genesis system config file and append to the list
							genesisSystemConfigPath := filepath.Join(parent, file.Name())
							genesisSystemConfig, err := os.ReadFile(genesisSystemConfigPath)
							if err != nil {
								panic(err)
							}
							var gsc GenesisSystemConfig
							err = json.Unmarshal(genesisSystemConfig, &gsc)
							if err != nil {
								panic(err)
							}
							gscMap[int(chain.ChainId)] = gsc
							generated++
							break
						}
					}
					if preGen == generated {
						fmt.Println("Failed to find genesis system config file:", name)
					}
				} else {
					fmt.Println("Skipping genesis config file without .json suffix:", name)
				}
			}
			fmt.Printf("Generated %d/%d genesis system configs for %s\n", generated, len(files), parent)
		} else {
			fmt.Printf("Skipping file %s in genesis-system-configs dir\n", dir.Name())
		}
	}

	// Only write the configs if the list is non-empty
	if len(gscMap) == 0 {
		fmt.Println("No genesis system configs found")
		return
	}

	// Write the genesis system configs to the output directory
	configs, err := json.MarshalIndent(gscMap, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filepath.Join(currentDir, "genesisSystemConfig.json"), configs, 0o644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Wrote genesis system configs to:", filepath.Join(currentDir, "genesisSystemConfig.json"))
}

type ChainEntry struct {
	Name            string   `json:"name" toml:"name"`
	Identifier      string   `json:"identifier" toml:"identifier"`
	ChainId         uint64   `json:"chainId" toml:"chain_id"`
	RPC             []string `json:"rpc" toml:"rpc"`
	Explorer        []string `json:"explorers" toml:"explorers"`
	SuperchainLevel uint     `json:"superchainLevel" toml:"superchain_level"`
	Parent          Parent   `json:"parent" toml:"parent"`
}

type Parent struct {
	Type    string   `json:"type" toml:"type"`
	Chain   string   `json:"chain" toml:"chain"`
	Bridges []string `json:"bridge,omitempty" toml:"bridges,omitempty"`
}

func currentDir() string {
	_, currentFilePath, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("Error: Unable to determine the current file path")
		os.Exit(1)
	}
	return filepath.Dir(currentFilePath)
}
