package deployer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

func TestNewOpDeployer(t *testing.T) {
	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))

	tests := []struct {
		name               string
		l1ContractsRelease string
		shouldError        bool
	}{
		{
			name:               "op-contracts/v1.6.0",
			l1ContractsRelease: "tag://op-contracts/v1.6.0",
			shouldError:        false,
		},
		{
			name:               "op-contracts/v2.0.0",
			l1ContractsRelease: "tag://op-contracts/v2.0.0",
			shouldError:        false,
		},
		{
			name:               "op-contracts/v3.0.0",
			l1ContractsRelease: "tag://op-contracts/v3.0.0",
			shouldError:        false,
		},
		{
			name:               "op-contracts/v4.0.0",
			l1ContractsRelease: "tag://op-contracts/v4.0.0",
			shouldError:        false,
		},
		{
			name:               "non-existent version",
			l1ContractsRelease: "tag://op-contracts/v999.999.999",
			shouldError:        true,
		},
		{
			name:               "empty release string",
			l1ContractsRelease: "",
			shouldError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			deployer, err := NewOpDeployer(lgr, tt.l1ContractsRelease)

			if tt.shouldError {
				require.Error(t, err)
				require.Nil(t, deployer)
			} else {
				require.NoError(t, err)
				require.NotNil(t, deployer)
				require.Equal(t, tt.l1ContractsRelease, deployer.l1ContractsRelease)
				require.NotEmpty(t, deployer.DeployerVersion)

				// Verify the binary exists
				homeDir, err := os.UserHomeDir()
				require.NoError(t, err)

				binaryPath := filepath.Join(homeDir, ".cache/binaries", deployer.DeployerVersion, "op-deployer")
				_, err = os.Stat(binaryPath)
				require.NoError(t, err, "Binary should exist at %s", binaryPath)
			}
		})
	}
}
