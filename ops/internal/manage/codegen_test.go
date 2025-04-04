package manage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/stretchr/testify/require"
)

func TestGenAddressesFile(t *testing.T) {
	require.NoError(t, GenAddressesFile("testdata"))

	addrsFile := paths.AddressesFile("testdata")
	t.Cleanup(func() {
		require.NoError(t, os.Remove(addrsFile))
	})

	expected, err := os.ReadFile("testdata/expected-addresses.json")
	require.NoError(t, err)

	actual, err := os.ReadFile(addrsFile)
	require.NoError(t, err)

	require.JSONEq(t, string(expected), string(actual))
}

func TestGenChainList(t *testing.T) {
	for _, ext := range []string{"toml", "json"} {
		t.Run(ext, func(t *testing.T) {
			t.Parallel()

			// Create a temporary directory
			tempDir := t.TempDir()
			outPath := filepath.Join(tempDir, fmt.Sprintf("chainList.%s", ext))

			require.NoError(t, GenChainListFile("testdata", outPath))
			actualBytes, err := os.ReadFile(outPath)
			require.NoError(t, err)
			expectedBytes, err := os.ReadFile(fmt.Sprintf("testdata/expected-chainList.%s", ext))
			require.NoError(t, err)
			require.Equal(t, strings.TrimSpace(string(expectedBytes)), strings.TrimSpace(string(actualBytes)))
		})
	}

	t.Run("any other extension", func(t *testing.T) {
		t.Parallel()

		err := GenChainListFile("testdata", "any-path.txt")
		require.ErrorContains(t, err, "unsupported file extension")
	})
}

func TestGenChainsReadme(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	readmeFile := filepath.Join(tempDir, "chains.md")

	require.NoError(t, GenChainsReadme("testdata", readmeFile))

	expectedBytes, err := os.ReadFile("testdata/expected-chains.md")
	require.NoError(t, err)

	actualBytes, err := os.ReadFile(readmeFile)
	require.NoError(t, err)

	require.Equal(t, strings.TrimSpace(string(expectedBytes)), strings.TrimSpace(string(actualBytes)))
}
