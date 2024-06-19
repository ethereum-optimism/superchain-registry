package main

import (
	"encoding/json"
	"errors"
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
	Proposer          = "Proposer"
	UnsafeBlockSigner = "UnsafeBlockSigner"
	BatchSubmitter    = "BatchSubmitter"

	// Non Fault Proof contracts
	L2OutputOracleProxy = "L2OutputOracleProxy"

	// Fault Proof contracts:
	AnchorStateRegistryProxy = "AnchorStateRegistryProxy"
	DelayedWETHProxy         = "DelayedWETHProxy"
	DisputeGameFactoryProxy  = "DisputeGameFactoryProxy"
	FaultDisputeGame         = "FaultDisputeGame"
	MIPS                     = "MIPS"
	PermissionedDisputeGame  = "PermissionedDisputeGame"
	PreimageOracle           = "PreimageOracle"
)

func readAddressesFromChain(addresses map[string]string, l1RpcUrl string, isFPAC bool) error {
	// SuperchainConfig
	address, err := castCall(addresses[OptimismPortalProxy], "superchainConfig()(address)", l1RpcUrl)
	if err != nil {
		addresses[SuperchainConfig] = ""
	} else {
		addresses[SuperchainConfig] = address
	}

	// Guardian
	address, err = castCall(addresses[SuperchainConfig], "guardian()(address)", l1RpcUrl)
	if err != nil {
		address, err = castCall(addresses[OptimismPortalProxy], "guardian()(address)", l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Guardian %w", err)
		}
	}
	addresses[Guardian] = address

	// ProxyAdminOwner
	address, err = castCall(addresses[ProxyAdmin], "owner()(address)", l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve address for ProxyAdminOwner")
	}
	addresses[ProxyAdminOwner] = address

	// SystemConfigOwner
	address, err = castCall(addresses[SystemConfigProxy], "owner()(address)", l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve address for SystemConfigOwner")
	}
	addresses[SystemConfigOwner] = address

	// UnsafeBlockSigner
	address, err = castCall(addresses[SystemConfigProxy], "unsafeBlockSigner()(address)", l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve address for UnsafeBlockSigner")
	}
	addresses[UnsafeBlockSigner] = address

	// BatchSubmitter
	hash, err := castCall(addresses[SystemConfigProxy], "batcherHash()(bytes32)", l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve batcherHash")
	}
	addresses[BatchSubmitter] = "0x" + hash[24:63]

	if isFPAC {
		// Proposer
		address, err = castCall(addresses[PermissionedDisputeGame], "proposer()(address)", l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Proposer")
		}
		addresses[UnsafeBlockSigner] = address

		// Challenger
		address, err = castCall(addresses[PermissionedDisputeGame], "challenger()(address)", l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Challenger")
		}
		addresses[Challenger] = address
	} else {
		// Proposer
		address, err = castCall(addresses[L2OutputOracleProxy], "PROPOSER()(address)", l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Proposer")
		}
		addresses[UnsafeBlockSigner] = address

		// Challenger
		address, err = castCall(addresses[L2OutputOracleProxy], "CHALLENGER()(address)", l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Challenger")
		}
		addresses[Challenger] = address
	}

	fmt.Printf("Addresses read from chain\n")
	return nil
}

func readAddressesFromJSON(contractAddresses map[string]string, deploymentsDir string) error {
	contractsFromJSON := []string{
		AddressManager,
		L1CrossDomainMessengerProxy,
		L1ERC721BridgeProxy,
		L1StandardBridgeProxy,
		OptimismMintableERC20FactoryProxy,
		SystemConfigProxy,
		OptimismPortalProxy,
		ProxyAdmin,
	}

	contractsFromJSONFPAC := append(contractsFromJSON, []string{
		AnchorStateRegistryProxy,
		DelayedWETHProxy,
		DisputeGameFactoryProxy,
		FaultDisputeGame,
		MIPS,
		PreimageOracle,
	}...)
	contractsFromJSONNonFPAC := append(contractsFromJSON, L2OutputOracleProxy)

	contracts := contractsFromJSONNonFPAC

	deployFilePath := filepath.Join(deploymentsDir, ".deploy")
	_, err := os.Stat(deployFilePath)

	if err != nil {
		// Use legacy deployment artifact schema

		_, err := os.ReadFile(filepath.Join(deploymentsDir, L2OutputOracleProxy+".json"))
		if errors.Is(err, os.ErrNotExist) {
			contracts = contractsFromJSONFPAC
		}
		for _, name := range contracts {
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

		_, err = addressList.AddressFor((L2OutputOracleProxy))
		if err != nil {
			contracts = contractsFromJSONFPAC
		}

		for _, name := range contracts {
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
