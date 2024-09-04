package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"

	"github.com/BurntSushi/toml"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
)

type ChainEntry struct {
	Name                 string   `json:"name" toml:"name"`
	Identifier           string   `json:"identifier" toml:"identifier"`
	ChainId              uint64   `json:"chainId" toml:"chain_id"`
	RPC                  []string `json:"rpc" toml:"rpc"`
	Explorer             []string `json:"explorers" toml:"explorers"`
	SuperchainLevel      uint     `json:"superchainLevel" toml:"superchain_level"`
	DataAvailabilityType string   `json:"dataAvailabilityType" toml:"data_availability_type"`
	Parent               Parent   `json:"parent" toml:"parent"`
	GasPayingToken       *Address `json:"gasPayingToken,omitempty" toml:"gas_paying_token,omitempty"`
}

type Parent struct {
	Type    string   `json:"type" toml:"type"`
	Chain   string   `json:"chain" toml:"chain"`
	Bridges []string `json:"bridge,omitempty" toml:"bridges,omitempty"`
}

func main() {
	allChains := make([]ChainEntry, 0)
	superchainTargets := make([]string, 0)
	chainAddresses := make(map[uint64]AddressList, 0)
	for t := range Superchains {
		superchainTargets = append(superchainTargets, t)
	}
	slices.Sort(superchainTargets)
	for _, sc := range superchainTargets {
		standardChains := make([]ChainEntry, 0)
		frontierChains := make([]ChainEntry, 0)
		for _, chainId := range Superchains[sc].ChainIDs {
			chain := OPChains[uint64(chainId)]
			if chain == nil {
				log.Fatalf("cannot find chain with id %d", chainId)
			}
			chainEntry := ChainEntry{
				Name:                 chain.Name,
				Identifier:           chain.Identifier(),
				ChainId:              chain.ChainID,
				RPC:                  []string{chain.PublicRPC},
				Explorer:             []string{chain.Explorer},
				SuperchainLevel:      uint(chain.SuperchainLevel),
				DataAvailabilityType: string(chain.DataAvailabilityType),
				Parent:               Parent{"L2", chain.Superchain, []string{}},
				GasPayingToken:       chain.GasPayingToken,
			}
			chainAddresses[chainId] = chain.Addresses
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

	currentDir := currentDir()
	repositoryRoot := filepath.Join(currentDir, "../../../")

	allChainsBytes, err := json.MarshalIndent(allChains, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filepath.Join(repositoryRoot, "chainList.json"), allChainsBytes, 0o644)
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
	err = os.WriteFile(filepath.Join(repositoryRoot, "chainList.toml"), buf.Bytes(), 0o644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Wrote chainList.toml file")

	// Write all chain configs to a single file
	type Superchain struct {
		Name         string           `json:"name" toml:"name"`
		Config       SuperchainConfig `json:"config" toml:"config"`
		ChainConfigs []ChainConfig    `json:"chains" toml:"chains"`
	}
	superchains := make([]Superchain, 0)
	for _, sc := range Superchains {
		chainConfigs := make([]ChainConfig, 0)
		for _, chainId := range sc.ChainIDs {
			chain, found := OPChains[uint64(chainId)]
			if !found {
				log.Fatalf("cannot find chain with id %d", chainId)
			}
			chainConfigs = append(chainConfigs, *chain)
		}
		sort.Slice(chainConfigs, func(i, j int) bool {
			return chainConfigs[i].ChainID < chainConfigs[j].ChainID
		})
		superchains = append(superchains, Superchain{
			Name:         sc.Superchain,
			Config:       sc.Config,
			ChainConfigs: chainConfigs,
		})
	}
	sort.Slice(superchains, func(i, j int) bool {
		return superchains[i].Config.Name < superchains[j].Config.Name
	})
	if len(superchains) != 0 {
		type FullSuperchains struct {
			Chains []Superchain `json:"superchains" toml:"superchains"`
		}
		superchainBytes, err := json.MarshalIndent(FullSuperchains{Chains: superchains}, "", "  ")
		if err != nil {
			panic(err)
		}

		// Write configs.json file to the top-level directory.
		err = os.WriteFile(filepath.Join(repositoryRoot, "superchain/configs/configs.json"), superchainBytes, 0o644)
		if err != nil {
			panic(err)
		}
		fmt.Println("Wrote configs.json file")
	}

	// Marshal to JSON
	addressesBytes, err := json.MarshalIndent(chainAddresses, "", "  ")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(filepath.Join(currentDir, "../../extra/addresses/addresses.json"), addressesBytes, 0o644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Wrote addresses.json file")
}

func currentDir() string {
	_, currentFilePath, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("Error: Unable to determine the current file path")
		os.Exit(1)
	}
	return filepath.Dir(currentFilePath)
}
