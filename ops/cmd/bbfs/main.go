package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Minimal struct to unmarshal relevant fields
type ChainConfig struct {
	Name      string `toml:"name"`
	ChainID   int    `toml:"chain_id"`
	Addresses struct {
		SystemConfigProxy string `toml:"SystemConfigProxy"`
	} `toml:"addresses"`
}

// ABI definition for the blobBaseFeeScalar function
const blobBaseFeeScalarABI = `[{"constant":true,"inputs":[],"name":"blobbasefeeScalar","outputs":[{"name":"","type":"uint32"}],"payable":false,"stateMutability":"view","type":"function"}]`

func main() {
	// RPC URL for Optimism Mainnet
	rpcURL := "https://ethereum-rpc.publicnode.com"

	// Connect to the Ethereum client
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to RPC: %v", err)
	}
	defer client.Close()

	// Relative path to the superchain registry
	relativePath := "../../../superchain/configs/mainnet"
	rootDir, err := filepath.Abs(relativePath)
	if err != nil {
		log.Fatalf("Error converting path to absolute: %v", err)
	}

	// Check if the directory exists
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		log.Fatalf("Error: directory does not exist: %s", rootDir)
	}

	fmt.Printf("Using Superchain Registry directory: %s\n", rootDir)
	fmt.Printf("%-30s | %-10s | %-42s | %-20s | %s \n", "Chain Name", "Chain ID", "SystemConfigProxy", "blobbasefeeScalar", "reverted")
	fmt.Println(strings.Repeat("-", 115))

	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(blobBaseFeeScalarABI))
	if err != nil {
		log.Fatalf("Failed to parse ABI: %v", err)
	}

	// Walk through the file tree
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing %s: %v\n", path, err)
			return nil
		}

		// Process only TOML files
		if filepath.Ext(path) == ".toml" {

			if filepath.Base(path) == "superchain.toml" {
				return nil // Skip the superchain.toml file
			}

			var config ChainConfig

			// Unmarshal the TOML content into the struct
			if _, err := toml.DecodeFile(path, &config); err != nil {
				log.Printf("Error unmarshalling TOML in %s: %v\n", path, err)
				return nil
			}

			// Validate address format
			if len(config.Addresses.SystemConfigProxy) != 42 || !strings.HasPrefix(config.Addresses.SystemConfigProxy, "0x") {
				log.Printf("Invalid address format in %s: %s\n", path, config.Addresses.SystemConfigProxy)
				return nil
			}

			// Convert the address to a common.Address type
			address := common.HexToAddress(config.Addresses.SystemConfigProxy)

			// Prepare the call to the blobBaseFeeScalar function

			callData, err := hexutil.Decode("0xec707517") // This is the method ID for blobbasefeeScalar
			if err != nil {
				log.Printf("Error decoding call data for %s: %v\n", address.Hex(), err)
				return nil
			}
			// Make the contract call
			var reverted bool
			result, err := client.CallContract(context.Background(), ethereum.CallMsg{
				To:   &address,
				Data: callData,
			}, nil)
			if err != nil {
				reverted = true
			}

			// Unpack the result
			var blobBaseFeeScalar uint32

			if !reverted {
				err = parsedABI.UnpackIntoInterface(&blobBaseFeeScalar, "blobbasefeeScalar", result)
				if err != nil {
					panic(err)
				}
			}

			// Print the result
			fmt.Printf("%-30s | %-10d | %-42s | %-20d | %t\n",
				config.Name, config.ChainID, config.Addresses.SystemConfigProxy, blobBaseFeeScalar, reverted)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error walking the file tree: %v\n", err)
	}

	fmt.Println("Done.")
}
