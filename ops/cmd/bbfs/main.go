package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Minimal struct to unmarshal relevant fields
type ChainConfig struct {
	Name      string `toml:"name"`
	ChainID   int    `toml:"chain_id"`
	Addresses struct {
		SystemConfigProxy string `toml:"SystemConfigProxy"`
	} `toml:"addresses"`
}

func main() {

	data := map[int]ChainConfig{}

	// Relative path to the superchain registry
	relativePath := "../../.."

	// Convert to an absolute path
	rootDir, err := filepath.Abs(relativePath)
	if err != nil {
		log.Fatalf("Error converting path to absolute: %v\n", err)
	}

	// Walk through the file tree
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing %s: %v\n", path, err)
			return nil
		}

		// Process only TOML files
		if filepath.Ext(path) == ".toml" {
			// Read the TOML file
			content, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Error reading file %s: %v\n", path, err)
				return nil
			}

			// Unmarshal the TOML content into the struct
			var config ChainConfig
			if err := toml.Unmarshal(content, &config); err != nil {
				log.Printf("Error unmarshalling TOML in %s: %v\n", path, err)
				return nil
			}

			// Skip files that don't have the required fields
			if config.Name == "" || config.ChainID == 0 || config.Addresses.SystemConfigProxy == "" {
				log.Printf("Skipping incomplete TOML file: %s\n", path)
				return nil
			}

			// Print the extracted information
			fmt.Printf("Chain Name: %s\nChain ID: %d\nSystemConfigProxy: %s\n",
				config.Name, config.ChainID, config.Addresses.SystemConfigProxy)
			fmt.Println("------------------------------------------------")

			data[config.ChainID] = config
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error walking the file tree: %v\n", err)
	}

	fmt.Println("Done.")
}
