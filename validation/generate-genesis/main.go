package main

import (
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/core"
)

func main() {
	rev := os.Args[1]

	for chainID, chain := range superchain.OPChains {
		gethGenesis, err := core.LoadOPStackGenesis(chainID)
		if err != nil {
			panic(err)
		}

		j, err := json.MarshalIndent(gethGenesis, "", "  ")
		if err != nil {
			panic(err)
		}

		dirPath := "./generate-genesis/output-" + rev

		// Create the directory if it doesn't exist
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			err := os.MkdirAll(dirPath, 0o755)
			if err != nil {
				log.Fatalf("Error creating directory: %v", err)
			}
		}

		err = os.WriteFile(path.Join(dirPath, chain.Chain+"-"+chain.Superchain+".json"), j, os.FileMode(0o644))
		if err != nil {
			panic(err)
		}
	}
}
