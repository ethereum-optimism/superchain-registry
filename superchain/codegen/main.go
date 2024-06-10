package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
)

type ChainEntry struct {
	Name            string `json:"name"`
	Identifier      string `json:"identifier"`
	ChainId         uint64 `json:"chainId"`
	PublicRPC       string `json:"public_rpc"`
	Explorer        string `json:"explorer"`
	SuperchainLevel uint   `json:"superchain_level"`
}

func main() {
	allChains := make([]ChainEntry, 0)
	for sc := range Superchains {
		standardChains := make([]ChainEntry, 0)
		frontierChains := make([]ChainEntry, 0)
		for _, chainId := range Superchains[sc].ChainIDs {
			chain := OPChains[uint64(chainId)]
			if chain == nil {
				log.Fatalf("cannot find chain with id %d", chainId)
			}
			chainEntry := ChainEntry{
				Name:            chain.Name,
				Identifier:      chain.Superchain, // + chain.shortName, // TODO
				ChainId:         chain.ChainID,
				PublicRPC:       chain.PublicRPC,
				Explorer:        chain.Explorer,
				SuperchainLevel: int(chain.SuperchainLevel),
			}
			switch chain.SuperchainLevel {
			case Standard:
				standardChains = append(standardChains, chainEntry)
			case Frontier:
				frontierChains = append(frontierChains, chainEntry)
			default:
				panic(fmt.Sprintf("unkown SuperchanLevel %d", chain.SuperchainLevel))

			}
		}
		allChains = append(allChains, standardChains...)
		allChains = append(allChains, frontierChains...)

	}
	allChainsBytes, err := json.MarshalIndent(allChains, "", "  ")
	if err != nil {
		panic(err)
	}

	// Determine the absolute path of the current file
	_, currentFilePath, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("Error: Unable to determine the current file path")
		os.Exit(1)
	}

	// Get the directory of the current file
	currentDir := filepath.Dir(currentFilePath)

	// Define the file path in the parent directory
	parentDir := filepath.Join(currentDir, "..")

	err = os.WriteFile(filepath.Join(parentDir, "chainList.json"), allChainsBytes, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Wrote chainList.json file")
}
