package manage

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	opgenesis "github.com/ethereum-optimism/optimism/op-core/genesis"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/stretchr/testify/require"
)

func TestWriteCompactGenesis(t *testing.T) {
	wd, err := paths.FindRepoRoot()
	require.NoError(t, err)
	inspected, err := os.ReadFile(filepath.Join("testdata", "expected-genesis.json"))
	require.NoError(t, err)

	genPath := filepath.Join(t.TempDir(), "genesis.json.zst")
	require.NoError(t, writeCompactGenesis(wd, genPath, inspected))

	written, err := ReadGenesis(wd, genPath)
	require.NoError(t, err)

	compact := new(bytes.Buffer)
	require.NoError(t, json.Compact(compact, inspected))
	require.Equal(t, compact.String(), string(written), "written genesis is not the compacted input")
	require.Equal(t, jsonTokens(t, inspected), jsonTokens(t, written), "written genesis changed the input's keys, values or key order")

	var genesis opgenesis.Genesis
	require.NoError(t, json.Unmarshal(written, &genesis))
}

// jsonTokens returns the JSON tokens of data in order, so that comparing them compares keys,
// values and key order.
func jsonTokens(t *testing.T, data []byte) []json.Token {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	var tokens []json.Token
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			return tokens
		}
		require.NoError(t, err)
		tokens = append(tokens, tok)
	}
}
