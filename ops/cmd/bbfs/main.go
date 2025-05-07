package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
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
	Name                 string `toml:"name"`
	ChainID              int    `toml:"chain_id"`
	DataAvailabilityType string `toml:"data_availability_type"`
	Addresses            struct {
		SystemConfigProxy string `toml:"SystemConfigProxy"`
	} `toml:"addresses"`
}

// ABI definition for the blobBaseFeeScalar function
const blobBaseFeeScalarABI = `[{"constant":true,"inputs":[],"name":"scalar","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`

// ABI definition for the version function
const versionABI = `[{"constant":true,"inputs":[],"name":"version","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"}]`

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

	// Parse the ABIs
	parsedABI, err := abi.JSON(strings.NewReader(blobBaseFeeScalarABI))
	if err != nil {
		log.Fatalf("Failed to parse scalar ABI: %v", err)
	}

	parsedVersionABI, err := abi.JSON(strings.NewReader(versionABI))
	if err != nil {
		log.Fatalf("Failed to parse version ABI: %v", err)
	}

	fmt.Printf("Using Superchain Registry directory: %s\n", rootDir)
	fmt.Printf("%-30s | %-10s | %-42s | %-20s | %-20s | %-10s | %-10s | %s\n",
		"Chain Name", "Chain ID", "SystemConfigProxy", "blobbasefeeScalar", "baseFeeScalar", "DA Type", "Scalar Ver", "Version")
	fmt.Println(strings.Repeat("-", 160))

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

			callData, err := hexutil.Decode("0xf45e65d8") // This is the method ID for scalar
			if err != nil {
				log.Printf("Error decoding call data for %s: %v\n", address.Hex(), err)
				return nil
			}
			// Make the contract calls
			var reverted bool
			result, err := client.CallContract(context.Background(), ethereum.CallMsg{
				To:   &address,
				Data: callData,
			}, nil)
			if err != nil {
				reverted = true
			}

			// Get version
			versionCallData, err := hexutil.Decode("0x54fd4d50") // This is the method ID for version()
			if err != nil {
				log.Printf("Error decoding version call data for %s: %v\n", address.Hex(), err)
				return nil
			}

			version := "unknown"
			versionResult, err := client.CallContract(context.Background(), ethereum.CallMsg{
				To:   &address,
				Data: versionCallData,
			}, nil)
			if err == nil {
				var versionStr string
				err = parsedVersionABI.UnpackIntoInterface(&versionStr, "version", versionResult)
				if err == nil {
					version = versionStr
				}
			}

			// Unpack the result
			scalarbig := new(big.Int)
			var es EcotoneScalars
			if !reverted {
				err = parsedABI.UnpackIntoInterface(&scalarbig, "scalar", result)
				if err != nil {
					panic(err)
				}

				scalar := [32]byte{}
				scalarbig.FillBytes(scalar[:])
				es, err = DecodeScalar(scalar)
				if err != nil {
					log.Printf("Error decoding scalar for %s: %v\n", address.Hex(), err)
					return nil
				}
			}

			if es.BlobBaseFeeScalar == 0 && config.DataAvailabilityType == "eth-da" {
				// Get scalar version string
				scalarVersion := "unknown"
				if !reverted {
					scalarbig := new(big.Int)
					err = parsedABI.UnpackIntoInterface(&scalarbig, "scalar", result)
					if err == nil {
						scalar := [32]byte{}
						scalarbig.FillBytes(scalar[:])
						if scalar[0] == L1ScalarBedrock {
							scalarVersion = "Bedrock"
						} else if scalar[0] == L1ScalarEcotone {
							scalarVersion = "Ecotone"
						}
					}
				}

				// Print the result
				fmt.Printf("%-30s | %-10d | %-42s | %-20d | %-20d | %-10s | %-10s | %s\n",
					config.Name, config.ChainID, config.Addresses.SystemConfigProxy,
					es.BlobBaseFeeScalar, es.BaseFeeScalar, config.DataAvailabilityType, scalarVersion, version)
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error walking the file tree: %v\n", err)
	}

	fmt.Println("Done.")
}

// DecodeScalar decodes the blobBaseFeeScalar and baseFeeScalar from a 32-byte scalar value.
// It uses the first byte to determine the scalar format.
func DecodeScalar(scalar [32]byte) (EcotoneScalars, error) {
	switch scalar[0] {
	case L1ScalarBedrock:
		return EcotoneScalars{
			BlobBaseFeeScalar: 0,
			BaseFeeScalar:     binary.BigEndian.Uint32(scalar[28:32]),
		}, nil
	case L1ScalarEcotone:
		return EcotoneScalars{
			BlobBaseFeeScalar: binary.BigEndian.Uint32(scalar[24:28]),
			BaseFeeScalar:     binary.BigEndian.Uint32(scalar[28:32]),
		}, nil
	default:
		return EcotoneScalars{}, fmt.Errorf("unexpected system config scalar: %x", scalar)
	}
}

type EcotoneScalars struct {
	BlobBaseFeeScalar uint32
	BaseFeeScalar     uint32
}

// The Ecotone upgrade introduces a versioned L1 scalar format
// that is backward-compatible with pre-Ecotone L1 scalar values.
const (
	// L1ScalarBedrock is implied pre-Ecotone, encoding just a regular-gas scalar.
	L1ScalarBedrock = byte(0)
	// L1ScalarEcotone is new in Ecotone, allowing configuration of both a regular and a blobs scalar.
	L1ScalarEcotone = byte(1)
)
