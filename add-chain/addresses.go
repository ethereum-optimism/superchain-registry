package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/superchain-registry/superchain"
)

type AddressData struct {
	Address string `json:"address"`
}

func readAddressesFromChain(addresses *superchain.AddressList, l1RpcUrl string, isFaultProofs bool) error {
	// SuperchainConfig
	address, err := castCall(addresses.OptimismPortalProxy, "superchainConfig()(address)", l1RpcUrl)
	if err == nil {
		addresses.SuperchainConfig = superchain.MustHexToAddress(address)
	}

	// Guardian
	address, err = castCall(addresses.SuperchainConfig, "guardian()(address)", l1RpcUrl)
	if err != nil {
		address, err = castCall(addresses.OptimismPortalProxy, "guardian()(address)", l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Guardian %w", err)
		}
	}
	addresses.Guardian = superchain.MustHexToAddress(address)

	// ProxyAdminOwner
	address, err = castCall(addresses.ProxyAdmin, "owner()(address)", l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve address for ProxyAdminOwner")
	}
	addresses.ProxyAdminOwner = superchain.MustHexToAddress(address)

	// SystemConfigOwner
	address, err = castCall(addresses.SystemConfigProxy, "owner()(address)", l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve address for SystemConfigOwner")
	}
	addresses.SystemConfigOwner = superchain.MustHexToAddress(address)

	// UnsafeBlockSigner
	address, err = castCall(addresses.SystemConfigProxy, "unsafeBlockSigner()(address)", l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve address for UnsafeBlockSigner")
	}
	addresses.UnsafeBlockSigner = superchain.MustHexToAddress(address)

	// BatchSubmitter
	hash, err := castCall(addresses.SystemConfigProxy, "batcherHash()(bytes32)", l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve batcherHash")
	}
	batchSubmitter := "0x" + hash[26:66]
	addresses.BatchSubmitter = superchain.MustHexToAddress(batchSubmitter)

	if isFaultProofs {
		// Proposer
		address, err = castCall(addresses.PermissionedDisputeGame, "proposer()(address)", l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Proposer")
		}
		addresses.Proposer = superchain.MustHexToAddress(address)

		// Challenger
		address, err = castCall(addresses.PermissionedDisputeGame, "challenger()(address)", l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Challenger")
		}
		addresses.Challenger = superchain.MustHexToAddress(address)
	} else {
		// Proposer
		address, err = castCall(addresses.L2OutputOracleProxy, "PROPOSER()(address)", l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Proposer")
		}
		addresses.Proposer = superchain.MustHexToAddress(address)

		// Challenger
		address, err = castCall(addresses.L2OutputOracleProxy, "CHALLENGER()(address)", l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Challenger")
		}
		addresses.Challenger = superchain.MustHexToAddress(address)
	}
	return nil
}

func readAddressesFromJSON(addressList *superchain.AddressList, deploymentsDir string) error {
	// Check for the following
	// 1. filepath == deploymentsDir
	// 2. filepath == deploymentsDir/.deploy
	// 3. filepath == deploymentsDir/<contract-name>.json
	deployFilePath := filepath.Join(deploymentsDir)
	fileInfo, err := os.Stat(deployFilePath)
	if err != nil {
		return fmt.Errorf("invalid deployment filepath provided: %w", err)
	}

	if fileInfo.IsDir() {
		deployFilePath = filepath.Join(deploymentsDir, ".deploy")
		_, err = os.Stat(deployFilePath)
		if err != nil {
			// Use legacy deployment artifact schema
			contractAddresses := make(map[string]string)
			fmt.Printf("failed to find .deploy file. Will look for legacy .json files")
			files, _ := os.ReadDir(deploymentsDir)
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				contractName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
				fileContents, err := os.ReadFile(filepath.Join(deploymentsDir, file.Name()))
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				var data AddressData
				if err = json.Unmarshal(fileContents, &data); err != nil {
					return fmt.Errorf("failed to unmarshal json: %w", err)
				}
				contractAddresses[contractName] = data.Address
				err = mapToAddressList(contractAddresses, addressList)
				if err != nil {
					return fmt.Errorf("failed to convert contracts map into AddressList: %w", err)
				}
			}
			return nil
		}
	}

	rawData, err := os.ReadFile(deployFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err = json.Unmarshal(rawData, &addressList); err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return nil
}

func mapToAddressList(m map[string]string, result *superchain.AddressList) error {
	out, err := toml.Marshal(m)
	if err != nil {
		return err
	}

	return toml.Unmarshal(out, result)
}
