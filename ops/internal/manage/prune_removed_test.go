package manage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

func writeRequiredSuperchainConfigs(t *testing.T, wd string) {
	t.Helper()
	for _, superchain := range requiredSuperchains {
		dir := paths.SuperchainDir(wd, superchain)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(paths.SuperchainConfig(wd, superchain), []byte(""), 0o644))
	}
}

func TestValidateRequiredSuperchains(t *testing.T) {
	wd := t.TempDir()
	writeRequiredSuperchainConfigs(t, wd)
	require.NoError(t, ValidateRequiredSuperchains(wd))

	require.NoError(t, os.Remove(paths.SuperchainConfig(wd, config.MainnetSuperchain)))
	err := ValidateRequiredSuperchains(wd)
	require.ErrorContains(t, err, "required superchain mainnet")
}

func TestPruneRemovedChains(t *testing.T) {
	wd := t.TempDir()
	writeRequiredSuperchainConfigs(t, wd)

	chainList := []config.ChainListEntry{{ChainID: 123}}
	chainListJSON, err := json.Marshal(chainList)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(paths.ChainListJsonFile(wd), chainListJSON, 0o644))

	addresses := config.AddressesJSON{"123": {}}
	addressesJSON, err := json.Marshal(addresses)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(paths.AddressesFile(wd)), 0o755))
	require.NoError(t, os.WriteFile(paths.AddressesFile(wd), addressesJSON, 0o644))

	lgr := log.NewLogger(log.DiscardHandler())
	require.NoError(t, PruneRemovedChains(lgr, wd))

	var gotChainList []config.ChainListEntry
	data, err := os.ReadFile(paths.ChainListJsonFile(wd))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &gotChainList))
	require.Empty(t, gotChainList)

	var gotAddresses config.AddressesJSON
	data, err = os.ReadFile(paths.AddressesFile(wd))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &gotAddresses))
	require.Empty(t, gotAddresses)
}
