package report

import (
	"os"
	"path"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestScanL2_2(t *testing.T) {
	l1RpcUrl := os.Getenv("SEPOLIA_RPC_URL")

	tests := []struct {
		name       string
		chainId    uint64
		statePath  string
		wantErr    string
		wantReport L2Report
	}{
		{
			name:      "successful scan state_v1",
			chainId:   uint64(1952805748),
			statePath: path.Join("testdata", "state_v1.json"),
			wantReport: L2Report{
				Release:             string(validation.Semver170),
				ProvidedGenesisHash: common.HexToHash("0x86da5aa4d6badbfcde840e0de7b239002c802ca0cfc2a97c7a964c3c5777cfc6"),
				StandardGenesisHash: common.HexToHash("0x86da5aa4d6badbfcde840e0de7b239002c802ca0cfc2a97c7a964c3c5777cfc6"),
				AccountDiffs:        []AccountDiff{},
			},
		},
		{
			name:      "successful scan state_v2",
			chainId:   uint64(336),
			statePath: path.Join("testdata", "state_v2.json"),
			wantReport: L2Report{
				Release:             string(validation.Semver170),
				ProvidedGenesisHash: common.HexToHash("0x49bbc2daef1e2d5e8c9bf233525e80a3e087e56d11cbae65cc7e2fbce2c1ee65"),
				StandardGenesisHash: common.HexToHash("0x49bbc2daef1e2d5e8c9bf233525e80a3e087e56d11cbae65cc7e2fbce2c1ee65"),
				AccountDiffs:        []AccountDiff{},
			},
		},
		{
			name:      "successful scan state_v3",
			chainId:   uint64(336),
			statePath: path.Join("testdata", "state_v3.json"),
			wantReport: L2Report{
				Release:             string(validation.Semver300),
				ProvidedGenesisHash: common.HexToHash("0x3496699b84c93be3963120ba6bbb00cae10780031a421dd9a83db752154f23e4"),
				StandardGenesisHash: common.HexToHash("0x3496699b84c93be3963120ba6bbb00cae10780031a421dd9a83db752154f23e4"),
				AccountDiffs:        []AccountDiff{},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			report, err := ScanL2_2(tt.statePath, tt.chainId, l1RpcUrl)

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
