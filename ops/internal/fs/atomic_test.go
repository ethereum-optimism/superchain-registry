package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAtomicWrite(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "atomic-test.txt")

	expData := []byte("hello, world")
	require.NoError(t, AtomicWrite(filename, 0o755, expData))

	// Verify that the data was written correctly
	actData, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.EqualValues(t, expData, actData)
}
