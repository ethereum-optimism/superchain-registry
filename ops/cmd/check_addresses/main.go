package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// addressRegex matches Ethereum addresses in the format 0x...
var addressRegex = regexp.MustCompile(`0x[0-9a-fA-F]{40}`)

// ContractData represents the structure of contract entries in the TOML file
type ContractData struct {
	Version               string `toml:"version"`
	Address               string `toml:"address,omitempty"`
	ImplementationAddress string `toml:"implementation_address,omitempty"`
}

// ReleaseData maps contract names to their ContractData
type ReleaseData map[string]ContractData

// VersionsData maps release versions to their respective ReleaseData
type VersionsData map[string]ReleaseData

func main() {
	// Parse command line arguments
	var (
		fixFlag  = flag.Bool("fix", false, "Fix incorrect addresses")
		allFlag  = flag.Bool("all", false, "Check all TOML files in validation directory")
		fileFlag = flag.String("file", "", "Check specific file")
	)
	flag.Parse()

	// Get the root directory of the repository
	rootDir, err := findRepoRoot()
	if err != nil {
		log.Fatalf("Error finding repository root: %v", err)
	}

	// Determine which files to check
	var filesToCheck []string
	if *allFlag {
		err := filepath.Walk(filepath.Join(rootDir, "validation"), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".toml") {
				filesToCheck = append(filesToCheck, path)
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Error walking validation directory: %v", err)
		}
	} else if *fileFlag != "" {
		filesToCheck = append(filesToCheck, *fileFlag)
	} else {
		// Default file
		filesToCheck = append(filesToCheck, filepath.Join(rootDir, "validation", "standard", "standard-versions-sepolia.toml"))
	}

	// Check each file
	totalIncorrect := 0
	for _, file := range filesToCheck {
		incorrectCount := checkFile(file, *fixFlag)
		totalIncorrect += incorrectCount
	}

	// Exit with error code if incorrect addresses were found and not fixed
	if totalIncorrect > 0 && !*fixFlag {
		os.Exit(1)
	}
}

// checkFile checks a single file for incorrectly checksummed addresses
func checkFile(tomlFile string, fix bool) int {
	// Read the TOML file content
	content, err := os.ReadFile(tomlFile)
	if err != nil {
		log.Fatalf("Error reading TOML file %s: %v", tomlFile, err)
	}

	// Original content as string
	originalContent := string(content)

	// Use regex to find all addresses
	addresses := addressRegex.FindAllString(originalContent, -1)

	// Map to keep track of incorrect addresses and their checksummed versions
	incorrectAddresses := make(map[string]string)

	// Check each address for correct checksum
	for _, addr := range addresses {
		checksumAddr := common.HexToAddress(addr).Hex()
		if addr != checksumAddr {
			incorrectAddresses[addr] = checksumAddr
			fmt.Printf("File %s: Incorrect address: %s should be %s\n", tomlFile, addr, checksumAddr)
		}
	}

	// If no incorrect addresses found, we're done
	if len(incorrectAddresses) == 0 {
		fmt.Printf("File %s: All addresses are correctly checksummed!\n", tomlFile)
		return 0
	}

	fmt.Printf("File %s: Found %d incorrectly checksummed addresses\n", tomlFile, len(incorrectAddresses))

	// Fix incorrect addresses if requested
	if fix {
		// Fix incorrect addresses in the content
		correctedContent := originalContent
		for incorrect, correct := range incorrectAddresses {
			correctedContent = strings.ReplaceAll(correctedContent, incorrect, correct)
		}

		// Write corrected content back to file
		err = os.WriteFile(tomlFile, []byte(correctedContent), 0644)
		if err != nil {
			log.Fatalf("Error writing corrected content to %s: %v", tomlFile, err)
		}

		fmt.Printf("File %s: Fixed %d addresses\n", tomlFile, len(incorrectAddresses))
	}

	return len(incorrectAddresses)
}

// findRepoRoot tries to find the repository root by looking for .repo-root file
func findRepoRoot() (string, error) {
	// Start from the current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// First try: check if we're already in the root (has .repo-root file)
	if _, err := os.Stat(filepath.Join(dir, ".repo-root")); err == nil {
		return dir, nil
	}

	// Second try: go up one level (we might be in the ops directory)
	parentDir := filepath.Dir(dir)
	if _, err := os.Stat(filepath.Join(parentDir, ".repo-root")); err == nil {
		return parentDir, nil
	}

	// If both fail, return the best guess (parent directory)
	return parentDir, nil
}
