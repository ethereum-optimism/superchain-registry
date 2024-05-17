package main

import (
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/superchain-registry/superchain"
)

func main() {
	rev := os.Args[1]

	for chainID, chain := range superchain.OPChains {
		config, err := rollup.LoadOPStackRollupConfig(chainID)
		if err != nil {
			panic(err)
		}
		j, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			panic(err)
		}

		dirPath := "./generate-rollup-config/output-" + rev

		// Create the directory if it doesn't exist
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			err := os.MkdirAll(dirPath, 0o755)
			if err != nil {
				log.Fatalf("Error creating directory: %v", err)
			}
		}

		err = os.WriteFile(path.Join(dirPath, chain.Superchain+"-"+chain.Name+".json"), j, os.FileMode(0o644))
		if err != nil {
			panic(err)
		}
	}
}
