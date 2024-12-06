package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation"
)

type AddressData struct {
	Address string `json:"address"`
}

func ReadAddressesFromChain(addresses *superchain.AddressList, l1RpcUrl string, isFaultProofs bool) error {
	// SuperchainConfig
	address, err := validation.CastCall(addresses.OptimismPortalProxy, "superchainConfig()(address)", nil, l1RpcUrl)
	if err == nil {
		addresses.SuperchainConfig = superchain.MustHexToAddress(address[0])
	}

	// Guardian
	address, err = validation.CastCall(addresses.SuperchainConfig, "guardian()(address)", nil, l1RpcUrl)
	if err != nil {
		address, err = validation.CastCall(addresses.OptimismPortalProxy, "guardian()(address)", nil, l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Guardian %w", err)
		}
	}
	addresses.Guardian = superchain.MustHexToAddress(address[0])

	// ProxyAdminOwner
	address, err = validation.CastCall(addresses.ProxyAdmin, "owner()(address)", nil, l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve address for ProxyAdminOwner")
	}
	addresses.ProxyAdminOwner = superchain.MustHexToAddress(address[0])

	// SystemConfigOwner
	address, err = validation.CastCall(addresses.SystemConfigProxy, "owner()(address)", nil, l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve address for SystemConfigOwner")
	}
	addresses.SystemConfigOwner = superchain.MustHexToAddress(address[0])

	// UnsafeBlockSigner
	address, err = validation.CastCall(addresses.SystemConfigProxy, "unsafeBlockSigner()(address)", nil, l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve address for UnsafeBlockSigner")
	}
	addresses.UnsafeBlockSigner = superchain.MustHexToAddress(address[0])

	// BatchSubmitter
	hash, err := validation.CastCall(addresses.SystemConfigProxy, "batcherHash()(bytes32)", nil, l1RpcUrl)
	if err != nil {
		return fmt.Errorf("could not retrieve batcherHash")
	}
	batchSubmitter := "0x" + hash[0][26:66]
	addresses.BatchSubmitter = superchain.MustHexToAddress(batchSubmitter)

	if isFaultProofs {
		// Proposer
		address, err = validation.CastCall(addresses.PermissionedDisputeGame, "proposer()(address)", nil, l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Proposer")
		}
		addresses.Proposer = superchain.MustHexToAddress(address[0])

		// Challenger
		address, err = validation.CastCall(addresses.PermissionedDisputeGame, "challenger()(address)", nil, l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Challenger")
		}
		addresses.Challenger = superchain.MustHexToAddress(address[0])
	} else {
		// Proposer
		address, err = validation.CastCall(addresses.L2OutputOracleProxy, "PROPOSER()(address)", nil, l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Proposer")
		}
		addresses.Proposer = superchain.MustHexToAddress(address[0])

		// Challenger
		address, err = validation.CastCall(addresses.L2OutputOracleProxy, "CHALLENGER()(address)", nil, l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Challenger")
		}
		addresses.Challenger = superchain.MustHexToAddress(address[0])
	}
	return nil
}

func ReadAddressesFromJSON(addressList *superchain.AddressList, deploymentsDir string) error {
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
