package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/BurntSushi/toml"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
)

type ChainEntry struct {
	Name            string   `json:"name" toml:"name"`
	Identifier      string   `json:"identifier" toml:"identifier"`
	ChainId         uint64   `json:"chainId" toml:"chainId"`
	RPC             []string `json:"rpc" toml:"rpc"`
	Explorer        []string `json:"explorers" toml:"explorers"`
	SuperchainLevel uint     `json:"superchain_level" toml:"superchain_level"`
	Parent          Parent   `json:"parent" toml:"parent"`
}

type Parent struct {
	Type    string   `json:"type" toml:"type"`
	Chain   string   `json:"chain" toml:"chain"`
	Bridges []string `json:"bridge,omitempty" toml:"bridges,omitempty"`
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
				RPC:             []string{chain.PublicRPC},
				Explorer:        []string{chain.Explorer},
				SuperchainLevel: uint(chain.SuperchainLevel),
				Parent:          Parent{"L2", chain.Superchain, []string{}},
			}
			switch chain.SuperchainLevel {
			case Standard:
				standardChains = append(standardChains, chainEntry)
			case Frontier:
				frontierChains = append(frontierChains, chainEntry)
			default:
				panic(fmt.Sprintf("unknown SuperchanLevel %d", chain.SuperchainLevel))
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
	parentDir := filepath.Join(currentDir, "../..")

	allChainsBytes, err := json.MarshalIndent(allChains, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filepath.Join(parentDir, "/configs/chainList.json"), allChainsBytes, 0o644)
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
	err = os.WriteFile(filepath.Join(parentDir, "/configs/chainList.toml"), buf.Bytes(), 0o644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Wrote chainList.toml file")

	addresses()
}

func addresses() {
	addressList := make(map[string]AddressList)
	for id := range OPChains {
		list := Addresses[id]
		addressList[strconv.Itoa(int(id))] = *list
	}
	addressListBytes, err := json.MarshalIndent(addressList, "", " ")
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
	path := filepath.Join(currentDir, "../../extra/addresses/addresses.json")

	err = os.WriteFile(path, addressListBytes, 0o644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Wrote addresses.json file")
}
