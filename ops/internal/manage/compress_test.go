package manage

import (
	"bytes"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisCompression(t *testing.T) {
	wd := "testdata"
	superchain := config.MainnetSuperchain
	shortName := "test"
	testGen := makeTestGenesis()

	testGenJSON, err := json.Marshal(testGen)
	require.NoError(t, err)

	require.NoError(t, WriteSuperchainGenesis(wd, superchain, shortName, testGenJSON))
	require.FileExists(t, paths.GenesisFile(wd, superchain, shortName))
	t.Cleanup(func() {
		require.NoError(t, os.Remove(paths.GenesisFile(wd, superchain, shortName)))
	})

	readGen, err := ReadSuperchainGenesis(wd, superchain, shortName)
	require.NoError(t, err)

	require.JSONEq(t, string(testGenJSON), string(readGen))
}

// TestGenesisRoundTrip checks that every genesis file in the registry reads and writes back to the
// same JSON bytes, so syncing a genesis never changes or drops any of its fields.
func TestGenesisRoundTrip(t *testing.T) {
	wd, err := paths.FindRepoRoot()
	require.NoError(t, err)

	genPaths, err := filepath.Glob(filepath.Join(paths.ExtraDir(wd), "genesis", "*", "*.json.zst"))
	require.NoError(t, err)
	require.NotEmpty(t, genPaths)

	for _, genPath := range genPaths {
		name, err := filepath.Rel(filepath.Join(paths.ExtraDir(wd), "genesis"), genPath)
		require.NoError(t, err)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gen, err := ReadGenesis(wd, genPath)
			require.NoError(t, err)

			outPath := filepath.Join(t.TempDir(), "genesis.json.zst")
			require.NoError(t, WriteGenesis(wd, outPath, gen))

			roundTripped, err := ReadGenesis(wd, outPath)
			require.NoError(t, err)
			require.True(t, bytes.Equal(gen, roundTripped), "genesis JSON changed in a write/read round trip")
		})
	}
}

func makeTestGenesis() *core.Genesis {
	testGenesis := &core.Genesis{
		Nonce:         0,
		Timestamp:     1722432480,
		ExtraData:     []byte("BEDROCK"),
		GasLimit:      30000000,
		Difficulty:    big.NewInt(2),
		Coinbase:      common.HexToAddress("0x4200000000000000000000000000000000000011"),
		Number:        2,
		GasUsed:       3,
		ParentHash:    common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
		BaseFee:       big.NewInt(4),
		ExcessBlobGas: new(uint64),
		BlobGasUsed:   new(uint64),
		Alloc: types.GenesisAlloc{
			common.Address{1}: {
				Code: []byte("CODE1"),
				Storage: map[common.Hash]common.Hash{
					common.HexToHash("0x01"): common.HexToHash("0x02"),
				},
				Balance: big.NewInt(0),
			},
			common.Address{2}: {
				Code:    []byte("CODE2"),
				Balance: big.NewInt(0),
			},
			common.Address{3}: {
				Storage: map[common.Hash]common.Hash{
					common.HexToHash("0x03"): common.HexToHash("0x04"),
				},
				Nonce:   2,
				Balance: big.NewInt(5),
			},
			common.Address{4}: {
				Nonce:   2,
				Balance: big.NewInt(0),
			},
		},
	}
	*testGenesis.ExcessBlobGas = 5
	*testGenesis.BlobGasUsed = 6
	return testGenesis
}
