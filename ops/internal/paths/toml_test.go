package paths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type tomlTestConfig struct {
	Name  string `toml:"name"`
	Inner struct {
		Known uint64 `toml:"known"`
	} `toml:"inner"`
}

func writeTOML(t *testing.T, content string) string {
	p := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	return p
}

func TestReadTOMLFile_RejectsUnknownKeys(t *testing.T) {
	p := writeTOML(t, "name = \"x\"\n[inner]\nknown = 1\nminBaseFee = 5\n")
	var out tomlTestConfig
	require.ErrorContains(t, ReadTOMLFile(p, &out), "inner.minBaseFee")
}

func TestReadTOMLFile_AcceptsModeledKeys(t *testing.T) {
	p := writeTOML(t, "name = \"x\"\n[inner]\nknown = 1\n")
	var out tomlTestConfig
	require.NoError(t, ReadTOMLFile(p, &out))
	require.Equal(t, "x", out.Name)
	require.EqualValues(t, 1, out.Inner.Known)
}
