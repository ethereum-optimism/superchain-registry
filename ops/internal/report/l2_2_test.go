package report

import (
	"os"
	"path"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/validation"
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
				Release:      string(validation.Semver170),
				GenesisDiffs: []string{},
			},
		},
		{
			name:      "successful scan state_v2",
			chainId:   uint64(336),
			statePath: path.Join("testdata", "state_v2.json"),
			wantReport: L2Report{
				Release:      string(validation.Semver170),
				GenesisDiffs: []string{},
			},
		},
		{
			name:      "successful scan state_v3",
			chainId:   uint64(336),
			statePath: path.Join("testdata", "state_v3.json"),
			wantReport: L2Report{
				Release:      string(validation.Semver300),
				GenesisDiffs: []string{},
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
