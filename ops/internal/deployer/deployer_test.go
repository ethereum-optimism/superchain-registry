package deployer

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

func TestAutodetectBinary(t *testing.T) {
	tests := []struct {
		name               string
		l1ContractsRelease string
		merger             StateMerger
		binPath            string
		shouldError        bool
	}{
		{
			name:               "op-contracts/v1.6.0",
			l1ContractsRelease: "tag://op-contracts/v1.6.0",
			merger:             MergeStateV1,
			binPath:            "op-deployer_v0.0.14",
			shouldError:        false,
		},
		{
			name:               "op-contracts/v2.0.0",
			l1ContractsRelease: "tag://op-contracts/v2.0.0",
			merger:             MergeStateV2,
			binPath:            "op-deployer_v0.2.3",
			shouldError:        false,
		},
		{
			name:               "op-contracts/v3.0.0",
			l1ContractsRelease: "tag://op-contracts/v3.0.0",
			merger:             MergeStateV3,
			binPath:            "op-deployer_v0.3.2",
			shouldError:        false,
		},
		{
			name:               "op-contracts/v4.0.0",
			l1ContractsRelease: "tag://op-contracts/v4.0.0",
			merger:             MergeStateV4_0,
			binPath:            "op-deployer_v0.4.2",
			shouldError:        false,
		},
		{
			name:               "op-contracts/v4.1.0",
			l1ContractsRelease: "tag://op-contracts/v4.1.0",
			merger:             MergeStateV4_1,
			binPath:            "op-deployer_v0.4.5",
			shouldError:        false,
		},
		{
			name:               "op-contracts/v5.0.0",
			l1ContractsRelease: "tag://op-contracts/v5.0.0",
			shouldError:        true,
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

	tmpDir := t.TempDir()
	for _, val := range contractVersions {
		stripped := strings.TrimPrefix(val, "op-deployer/")
		fp := path.Join(tmpDir, fmt.Sprintf("op-deployer_%s", stripped))
		_, err := os.Create(fp)
		require.NoError(t, err)
		require.NoError(t, os.Chmod(fp, 0o755))
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			picker, err := WithReleaseBinary(tmpDir, tt.l1ContractsRelease)

			if tt.shouldError {
				require.Error(t, err)
				require.Nil(t, picker)
			} else {
				require.NoError(t, err)
				require.NotNil(t, picker)

				require.Equal(t, reflect.ValueOf(tt.merger).Pointer(), reflect.ValueOf(picker.Merger()).Pointer())
				require.Equal(t, tt.binPath, filepath.Base(picker.Path()))
			}
		})
	}
}

func TestVersionsMapInitialization(t *testing.T) {
	// Verify the map is not empty
	if len(contractVersions) == 0 {
		t.Error("contractVersions map is empty, expected it to be populated from versions.json")
	}

	// Test a known key-value pair from versions.json
	expectedVersion := "op-deployer/v0.0.14"
	actualVersion, exists := contractVersions["op-contracts/v1.6.0"]

	require.True(t, exists, "expected key 'op-contracts/v1.6.0' not found in contractVersions map")
	require.Equal(t, actualVersion, expectedVersion)
}

func TestBinaryInvocation(t *testing.T) {
	cacheDir := os.Getenv("DEPLOYER_CACHE_DIR")
	require.NotEmpty(t, cacheDir)
	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, false))

	picker, err := WithReleaseBinary(cacheDir, "tag://op-contracts/v1.6.0")
	require.NoError(t, err)

	deployer, err := NewOpDeployer(lgr, picker)
	require.NoError(t, err)

	output, err := deployer.runCommand("--help")
	require.NoError(t, err)
	require.NotNil(t, output)
}
