package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum-optimism/superchain-registry/superchain"
)

type AddressData struct {
	Address string `json:"address"`
}

var (
	// Addresses to retrieve from JSON
	AddressManager                    = "AddressManager"
	L1CrossDomainMessengerProxy       = "L1CrossDomainMessengerProxy"
	L1ERC721BridgeProxy               = "L1ERC721BridgeProxy"
	L1StandardBridgeProxy             = "L1StandardBridgeProxy"
	L2OutputOracleProxy               = "L2OutputOracleProxy"
	OptimismMintableERC20FactoryProxy = "OptimismMintableERC20FactoryProxy"
	SystemConfigProxy                 = "SystemConfigProxy"
	OptimismPortalProxy               = "OptimismPortalProxy"
	ProxyAdmin                        = "ProxyAdmin"

	// Addresses to retrieve from chain
	SuperchainConfig  = "SuperchainConfig"
	Guardian          = "Guardian"
	Challenger        = "Challenger"
	ProxyAdminOwner   = "ProxyAdminOwner"
	SystemConfigOwner = "SystemConfigOwner"
)

func readAddressesFromChain(contractAddresses map[string]string, l1RpcUrl string) error {
	// SuperchainConfig
	address, err := executeCommand("cast", []string{"call", contractAddresses[OptimismPortalProxy], "superchainConfig()(address)", "-r", l1RpcUrl})
	address = strings.Join(strings.Fields(address), "") // remove whitespace
	if err != nil || address == "" || address == "0x" {
		contractAddresses[SuperchainConfig] = ""
	} else {
		contractAddresses[SuperchainConfig] = address
	}

	// Guardian
	address, err = executeCommand("cast", []string{"call", contractAddresses[SuperchainConfig], "guardian()(address)", "-r", l1RpcUrl})
	address = strings.Join(strings.Fields(address), "") // remove whitespace
	if err != nil || address == "" || address == "0x" {
		address, err = executeCommand("cast", []string{"call", contractAddresses[OptimismPortalProxy], "GUARDIAN()(address)", "-r", l1RpcUrl})
		address = strings.Join(strings.Fields(address), "") // remove whitespace
		if err != nil || address == "" || address == "0x" {
			return fmt.Errorf("could not retrieve address for Guardian")
		}
		contractAddresses[Guardian] = address
	} else {
		contractAddresses[Guardian] = address
	}

	// Challenger
	address, err = executeCommand("cast", []string{"call", contractAddresses[L2OutputOracleProxy], "challenger()(address)", "-r", l1RpcUrl})
	address = strings.Join(strings.Fields(address), "") // remove whitespace
	if err != nil || address == "" || address == "0x" {
		return fmt.Errorf("could not retrieve address for Guardian")
	} else {
		contractAddresses[Challenger] = address
	}

	// ProxyAdminOwner
	address, err = executeCommand("cast", []string{"call", contractAddresses[ProxyAdmin], "owner()(address)", "-r", l1RpcUrl})
	address = strings.Join(strings.Fields(address), "") // remove whitespace
	if err != nil || address == "" || address == "0x" {
		return fmt.Errorf("could not retrieve address for ProxyAdminOwner")
	} else {
		contractAddresses[ProxyAdminOwner] = address
	}

	// SystemConfigOwner
	address, err = executeCommand("cast", []string{"call", contractAddresses[SystemConfigProxy], "owner()(address)", "-r", l1RpcUrl})
	address = strings.Join(strings.Fields(address), "") // remove whitespace
	if err != nil || address == "" || address == "0x" {
		return fmt.Errorf("could not retrieve address for ProxyAdminOwner")
	} else {
		contractAddresses[SystemConfigOwner] = address
	}
	fmt.Printf("Contract addresses read from on-chain contracts\n")

	return nil
}

func readAddressesFromJSON(contractAddresses map[string]string, deploymentsDir string) error {
	var contractsFromJSON = []string{
		AddressManager,
		L1CrossDomainMessengerProxy,
		L1ERC721BridgeProxy,
		L1StandardBridgeProxy,
		L2OutputOracleProxy,
		OptimismMintableERC20FactoryProxy,
		SystemConfigProxy,
		OptimismPortalProxy,
		ProxyAdmin,
	}

	deployFilePath := filepath.Join(deploymentsDir, ".deploy")
	_, err := os.Stat(deployFilePath)

	if err != nil {
		// Use legacy deployment artifact schema
		for _, name := range contractsFromJSON {
			path := filepath.Join(deploymentsDir, name+".json")
			file, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file: %v", err)
			}
			var data AddressData
			if err = json.Unmarshal(file, &data); err != nil {
				return fmt.Errorf("failed to unmarshal json: %v", err)
			}
			contractAddresses[name] = data.Address
			fmt.Printf("%s : %s\n", name, data.Address)
		}
	} else {
		var addressList superchain.AddressList
		rawData, err := os.ReadFile(deployFilePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}

		if err = json.Unmarshal(rawData, &addressList); err != nil {
			return fmt.Errorf("failed to unmarshal json: %v", err)
		}

		for _, name := range contractsFromJSON {
			address, err := addressList.AddressFor(name)
			if err != nil {
				return fmt.Errorf("failed to retrieve %s address from list: %v", name, err)
			}
			contractAddresses[name] = address.String()
		}
	}

	fmt.Printf("Contract addresses read from deployments directory: %s\n", deploymentsDir)
	return nil
}

func writeAddressesToJSON(contractsAddresses map[string]string, superchainRepoPath, target, chainName string) error {
	dirPath := filepath.Join(superchainRepoPath, "superchain", "extra", "addresses", target)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	filePath := filepath.Join(dirPath, chainName+".json")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Marshal the map to JSON
	jsonData, err := json.MarshalIndent(contractsAddresses, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	// Write the JSON data to the file
	if _, err := file.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write json to file: %w", err)
	}
	fmt.Printf("Contract addresses written to: %s\n", filePath)

	return nil
}
