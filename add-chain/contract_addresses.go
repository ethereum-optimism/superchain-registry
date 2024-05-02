package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
	address, err := castCall(contractAddresses[OptimismPortalProxy], "superchainConfig()(address)", l1RpcUrl)
	if err != nil {
		contractAddresses[SuperchainConfig] = ""
	} else {
		contractAddresses[SuperchainConfig] = address
	}

	// Guardian
	address, err = castCall(contractAddresses[SuperchainConfig], "guardian()(address)", l1RpcUrl)
	if err != nil {
		address, err = castCall(contractAddresses[OptimismPortalProxy], "guardian()(address)", l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Guardian %w", err)
		}
	}
	contractAddresses[Guardian] = address

	// Challenger
	address, err = castCall(contractAddresses[L2OutputOracleProxy], "challenger()(address)", l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve address for Guardian")
	}
	contractAddresses[Challenger] = address

	// ProxyAdminOwner
	address, err = castCall(contractAddresses[ProxyAdmin], "owner()(address)", l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve address for ProxyAdminOwner")
	}
	contractAddresses[ProxyAdminOwner] = address

	// SystemConfigOwner
	address, err = castCall(contractAddresses[SystemConfigProxy], "owner()(address)", l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve address for ProxyAdminOwner")
	}
	contractAddresses[SystemConfigOwner] = address

	fmt.Printf("Contract addresses read from on-chain contracts\n")
	return nil
}

func readAddressesFromJSON(contractAddresses map[string]string, deploymentsDir string) error {
	contractsFromJSON := []string{
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
				return fmt.Errorf("failed to read file: %w", err)
			}
			var data AddressData
			if err = json.Unmarshal(file, &data); err != nil {
				return fmt.Errorf("failed to unmarshal json: %w", err)
			}
			contractAddresses[name] = data.Address
		}
	} else {
		var addressList superchain.AddressList
		rawData, err := os.ReadFile(deployFilePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		if err = json.Unmarshal(rawData, &addressList); err != nil {
			return fmt.Errorf("failed to unmarshal json: %w", err)
		}

		for _, name := range contractsFromJSON {
			address, err := addressList.AddressFor(name)
			if err != nil {
				return fmt.Errorf("failed to retrieve %s address from list: %w", name, err)
			}
			contractAddresses[name] = address.String()
		}
	}

	fmt.Printf("Contract addresses read from deployments directory: %s\n", deploymentsDir)
	return nil
}

func writeAddressesToJSON(contractsAddresses map[string]string, superchainRepoPath, target, chainName string) error {
	dirPath := filepath.Join(superchainRepoPath, "superchain", "extra", "addresses", target)
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
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
