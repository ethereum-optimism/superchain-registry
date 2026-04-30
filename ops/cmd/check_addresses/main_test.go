package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressRegex(t *testing.T) {
	// Test that the regex matches Ethereum addresses correctly
	validAddresses := []string{
		"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
		"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed",
		"0x0000000000000000000000000000000000000000",
		"0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
	}

	for _, addr := range validAddresses {
		match := addressRegex.MatchString(addr)
		assert.True(t, match, "Address should match: %s", addr)
	}

	invalidAddresses := []string{
		"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeA",     // Too short
		"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAedFF", // Too long
		"5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",     // Missing 0x prefix
		"0xGAAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",   // Invalid character
	}

	for _, addr := range invalidAddresses {
		match := addressRegex.MatchString(addr)
		assert.False(t, match, "Address should not match: %s", addr)
	}
}

func TestFindRepoRoot(t *testing.T) {
	// Create a temporary test directory structure
	tmpDir, err := os.MkdirTemp("", "check_addresses_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a .repo-root file in the temp directory
	repoRootFile := filepath.Join(tmpDir, ".repo-root")
	err = os.WriteFile(repoRootFile, []byte("test repo root"), 0644)
	assert.NoError(t, err)

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "ops")
	err = os.Mkdir(subDir, 0755)
	assert.NoError(t, err)

	// Test findRepoRoot from the tmpDir
	oldWd, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	assert.NoError(t, err)

	root, err := findRepoRoot()
	assert.NoError(t, err)
	assert.Equal(t, tmpDir, root)

	// Test findRepoRoot from the subDir
	err = os.Chdir(subDir)
	assert.NoError(t, err)

	root, err = findRepoRoot()
	assert.NoError(t, err)
	assert.Equal(t, tmpDir, root)
}

func TestCheckFile(t *testing.T) {
	// Create a temporary copy of our test file
	tmpDir, err := os.MkdirTemp("", "check_addresses_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Ensure testdata directory exists
	testdataDir := filepath.Join("testdata")
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		err = os.Mkdir(testdataDir, 0755)
		assert.NoError(t, err)
	}

	// Define paths
	srcFilePath := filepath.Join("testdata", "test_addresses.toml")
	tmpFilePath := filepath.Join(tmpDir, "test_addresses.toml")

	// Create the test file if it doesn't exist
	if _, err := os.Stat(srcFilePath); os.IsNotExist(err) {
		testContent := `# Test file for address validation

# Incorrect addresses (all lowercase)
[version1.contracts.contract1]
version = "1.0.0"
address = "0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed"

[version1.contracts.contract2]
version = "1.0.0"
implementation_address = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

# Correct addresses (checksummed)
[version2.contracts.contract1]
version = "2.0.0"
address = "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"

[version2.contracts.contract2]
version = "2.0.0"
implementation_address = "0xaAaAaAaaAaAaAaaAaAAAAAAAAaaaAaAaAaaAaaAa"`
		err = os.WriteFile(srcFilePath, []byte(testContent), 0644)
		assert.NoError(t, err)
	}

	// Copy test file to temp directory
	content, err := os.ReadFile(srcFilePath)
	assert.NoError(t, err)
	err = os.WriteFile(tmpFilePath, content, 0644)
	assert.NoError(t, err)

	// Test checking without fixing
	incorrectCount := checkFile(tmpFilePath, false)
	assert.Equal(t, 2, incorrectCount, "Should find two incorrect addresses")

	// Test checking with fixing
	incorrectCount = checkFile(tmpFilePath, true)
	assert.Equal(t, 2, incorrectCount, "Should find and fix two incorrect addresses")

	// Verify fix was applied
	incorrectCount = checkFile(tmpFilePath, false)
	assert.Equal(t, 0, incorrectCount, "Should find no incorrect addresses after fixing")

	// Read the fixed file and verify the content
	fixedContent, err := os.ReadFile(tmpFilePath)
	assert.NoError(t, err)
	fixedContentStr := string(fixedContent)

	// Check if the specific addresses were fixed
	assert.Contains(t, fixedContentStr, "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed")
	assert.NotContains(t, fixedContentStr, "0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	assert.Contains(t, fixedContentStr, "0xaAaAaAaaAaAaAaaAaAAAAAAAAaaaAaAaAaaAaaAa")
	assert.NotContains(t, fixedContentStr, "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
}
