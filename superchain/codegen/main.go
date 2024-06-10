package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
)

type ChainEntry struct {
	Name            string `json:"name" toml:"name"`
	Identifier      string `json:"identifier" toml:"identifier"`
	ChainId         uint64 `json:"chainId" toml:"chainId"`
	PublicRPC       string `json:"public_rpc" toml:"public_rpc"`
	Explorer        string `json:"explorer" toml:"explorer"`
	SuperchainLevel uint   `json:"superchain_level" toml:"superchain_level"`
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
				Identifier:      chain.Identifier(),
				ChainId:         chain.ChainID,
				PublicRPC:       chain.PublicRPC,
				Explorer:        chain.Explorer,
				SuperchainLevel: uint(chain.SuperchainLevel),
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

	fmt.Printf("Found %d chains...\n", len(allChains))

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

	allChainsBytes, err := json.MarshalIndent(allChains, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filepath.Join(parentDir, "chainList.json"), allChainsBytes, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Wrote chainList.json file")

	var buf bytes.Buffer
	allChainsForTOML := map[string]([]ChainEntry){"chains": allChains}
	if err := toml.NewEncoder(&buf).Encode(allChainsForTOML); err != nil {
		fmt.Println("Error encoding TOML:", err)
		return
	}
	err = os.WriteFile(filepath.Join(parentDir, "chainList.toml"), buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Wrote chainList.toml file")
}
