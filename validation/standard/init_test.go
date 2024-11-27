package standard

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigInitialization(t *testing.T) {
	// Ensure Config is initialized
	require.NotNil(t, Config, "Config should not be nil")
	require.NotNil(t, Config.Params, "Config.Params should not be nil")
	require.NotNil(t, Config.Roles, "Config.Roles should not be nil")
	require.NotNil(t, Config.MultisigRoles, "Config.MultisigRoles should not be nil")

	// Check individual network configurations
	networks := []string{"mainnet", "sepolia"}
	for _, network := range networks {
		t.Run(fmt.Sprintf("Params[%s]", network), func(t *testing.T) {
			// Ensure network Params are populated
			require.NotNil(t, Config.Params[network], "Config.Params[%s] should not be nil", network)
			require.NoError(t, Config.Params[network].Check(), "Config.Params[%s] has invalid zero value", network)
		})

		t.Run(fmt.Sprintf("MultisigRoles[%s]", network), func(t *testing.T) {
			// Ensure network MultisigRoles are populated
			require.NotNil(t, Config.MultisigRoles[network], "Config.MultisigRoles[%s] should not be nil", network)
			require.NotZero(t, Config.MultisigRoles[network], "Config.MultisigRoles[%s] should not be zero value", network)
		})

		t.Run(fmt.Sprintf("NetworkVersion[%s]", network), func(t *testing.T) {
			// Ensure network Versions are populated
			versions, ok := NetworkVersions[network]
			require.True(t, ok, "NetworkVersions[%s] should exist", network)
			require.NotNil(t, versions, "NetworkVersions[%s] should not be nil", network)
			require.NotZero(t, len(versions.Releases), "NetworkVersions[%s].Releases should not be empty", network)
		})
	}

	t.Run("Release", func(t *testing.T) {
		require.NotEmpty(t, Release, "Release should not be empty")
	})

	t.Run("BytecodeHashes", func(t *testing.T) {
		require.NotNil(t, BytecodeHashes[Release], "BytecodeHashes should not be nil")
		require.NotZero(t, len(BytecodeHashes), "BytecodeHashes should not be empty")
	})

	t.Run("BytecodeImmutables", func(t *testing.T) {
		require.NotNil(t, BytecodeImmutables[Release], "BytecodeImmutables should not be nil")
		require.NotZero(t, len(BytecodeImmutables), "BytecodeImmutables should not be empty")
	})
}
