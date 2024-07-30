package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
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

func readAddressesFromChain(addresses map[string]string, l1RpcUrl string, isFaultProofs bool) error {
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
	addresses[BatchSubmitter] = "0x" + hash[26:66]

	if isFaultProofs {
		// Proposer
		address, err = castCall(addresses[PermissionedDisputeGame], "proposer()(address)", l1RpcUrl)
		if err != nil {
			return fmt.Errorf("could not retrieve address for Proposer")
		}
		addresses[Proposer] = address

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
		addresses[Proposer] = address

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
	universalContracts := []string{
		AddressManager,
		L1CrossDomainMessengerProxy,
		L1ERC721BridgeProxy,
		L1StandardBridgeProxy,
		OptimismMintableERC20FactoryProxy,
		SystemConfigProxy,
		OptimismPortalProxy,
		ProxyAdmin,
	}

	customGasTokenContracts := []string{
		AnchorStateRegistryProxy,
		DelayedWETHProxy,
		DisputeGameFactoryProxy,
		MIPS,
		PreimageOracle,
		L2OutputOracleProxy,
	}

	faultProofContracts := []string{
		AnchorStateRegistryProxy,
		DelayedWETHProxy,
		DisputeGameFactoryProxy,
		FaultDisputeGame,
		MIPS,
		PermissionedDisputeGame,
		PreimageOracle,
	}

	contracts := universalContracts

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
			fmt.Printf("failed to find .deploy file. Will look for legacy .json files")

			if _, err := os.Stat(filepath.Join(deploymentsDir, FaultDisputeGame+".json")); err == nil {
				contracts = append(contracts, faultProofContracts...)
			} else if _, err := os.Stat(filepath.Join(deploymentsDir, AnchorStateRegistryProxy+".json")); err == nil {
				contracts = append(contracts, customGasTokenContracts...)
			} else {
				contracts = append(contracts, L2OutputOracleProxy)
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
			return nil
		}
	}

	var addressList superchain.AddressList
	rawData, err := os.ReadFile(deployFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err = json.Unmarshal(rawData, &addressList); err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	if _, err = addressList.AddressFor(FaultDisputeGame); err == nil {
		contracts = append(contracts, faultProofContracts...)
	} else if _, err = addressList.AddressFor(AnchorStateRegistryProxy); err == nil {
		contracts = append(contracts, customGasTokenContracts...)
	} else {
		contracts = append(contracts, L2OutputOracleProxy)
	}

	for _, name := range contracts {
		address, err := addressList.AddressFor(name)
		if err != nil {
			return fmt.Errorf("failed to retrieve %s address from list: %w", name, err)
		}
		contractAddresses[name] = address.String()
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
