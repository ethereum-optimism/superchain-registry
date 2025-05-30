package report

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/core"
	"github.com/stretchr/testify/require"
)

func TestScanL2(t *testing.T) {
	chainCfgData := readTestData(t, "chain-config.toml")
	genesisData := readTestData(t, "genesis.json.gz")
	statePath := path.Join("testdata", "deployer-state.json")
	// TODO: remove this hardcoded value if possible
	l1RpcUrl := "https://ci-sepolia-l1-archive.optimism.io"

	tests := []struct {
		name       string
		setup      func(*config.StagedChain, *core.Genesis)
		wantErr    string
		wantReport L2Report
	}{
		{
			name: "successful scan",
			setup: func(*config.StagedChain, *core.Genesis) {
			},
			wantReport: L2Report{
				Release:      string(validation.Semver170),
				GenesisDiffs: []string{},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var chainCfg config.StagedChain
			require.NoError(t, toml.Unmarshal(chainCfgData, &chainCfg))
			gzr, err := gzip.NewReader(bytes.NewReader(genesisData))
			require.NoError(t, err)
			var genesis deployer.OpaqueMap
			require.NoError(t, json.NewDecoder(gzr).Decode(&genesis))
			require.NoError(t, gzr.Close())

			var gen core.Genesis
			tt.setup(&chainCfg, &gen)
			report, err := ScanL2(statePath, chainCfg.ChainID, l1RpcUrl)

			if tt.wantErr == "" {
				require.NoError(t, err)
				require.Equal(t, tt.wantReport, *report)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
				require.Nil(t, report)
			}
		})
	}
}

func readTestData(t *testing.T, fname string) []byte {
	data, err := os.ReadFile(path.Join("testdata", fname))
	require.NoError(t, err)
	return data
}
