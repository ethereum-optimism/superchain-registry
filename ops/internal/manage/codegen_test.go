package manage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch/script"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

func TestGenChainsReadme(t *testing.T) {
	readmeFile, err := os.CreateTemp("", "chains-*.md")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Remove(readmeFile.Name()))
	})

	require.NoError(t, GenChainsReadme("testdata", readmeFile.Name()))

	expectedBytes, err := os.ReadFile("testdata/CHAINS.md")
	require.NoError(t, err)

	actualBytes, err := os.ReadFile(readmeFile.Name())
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(string(expectedBytes)), strings.TrimSpace(string(actualBytes)))
}

// loadTestAddressesJSON loads the expected addresses JSON from testdata
func loadTestAddressesJSON(t *testing.T) config.AddressesJSON {
	data, err := os.ReadFile(paths.AddressesFile("testdata"))
	require.NoError(t, err)

	var addresses config.AddressesJSON
	err = json.Unmarshal(data, &addresses)
	require.NoError(t, err)

	return addresses
}

// loadTestChainList loads the expected chain list from testdata
func loadTestChainList(t *testing.T) []config.ChainListEntry {
	data, err := os.ReadFile(filepath.Join("testdata", "chainList.json"))
	require.NoError(t, err)

	var chainList []config.ChainListEntry
	err = json.Unmarshal(data, &chainList)
	require.NoError(t, err)

	return chainList
}

// createTestChainConfigs creates chain configs for testing based on testdata
func createTestChainConfigs(t *testing.T) map[uint64]script.ChainConfig {
	addresses := loadTestAddressesJSON(t)
	chainList := loadTestChainList(t)

	chainCfgs := make(map[uint64]script.ChainConfig)
	for _, entry := range chainList {
		chainID := entry.ChainID
		chainIDStr := fmt.Sprintf("%d", chainID)

		// Only include chains that exist in addresses
		if chainAddrs, ok := addresses[chainIDStr]; ok {
			chainCfgs[chainID] = script.ChainConfig{
				Addresses:        chainAddrs.Addresses,
				Roles:            chainAddrs.Roles,
				FaultProofStatus: entry.FaultProofStatus,
			}
		}
	}

	return chainCfgs
}

func TestCodegenSyncer_NewCodegenSyncer(t *testing.T) {
	chainCfgs := createTestChainConfigs(t)

	// Test successful initialization
	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelError, false))
	syncer, err := NewCodegenSyncer(lgr, "./testdata", chainCfgs)

	require.NoError(t, err)
	require.NotNil(t, syncer)
	require.NotEmpty(t, syncer.ChainList)
	require.NotEmpty(t, syncer.Addresses)
	require.Equal(t, "./testdata", syncer.inputWd)
	require.Equal(t, "./testdata", syncer.outputWd)

	// Test initialization with separate output directory
	tempDir := t.TempDir()
	syncer, err = NewCodegenSyncer(lgr, "./testdata", chainCfgs, WithOutputDirectory(tempDir))
	require.NoError(t, err)
	require.Equal(t, "./testdata", syncer.inputWd)
	require.Equal(t, tempDir, syncer.outputWd)

	// Test initialization with invalid directory
	_, err = NewCodegenSyncer(lgr, "/nonexistent", chainCfgs)
	require.Error(t, err)
}

func TestCodegenSyncer_UpdateChainList(t *testing.T) {
	chainCfgs := createTestChainConfigs(t)
	var testChainID uint64
	for id := range chainCfgs {
		testChainID = id
		break
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelError, false))
	syncer, err := NewCodegenSyncer(lgr, "testdata", chainCfgs)
	require.NoError(t, err)

	err = syncer.UpdateChainList(fmt.Sprintf("%d", testChainID), script.ChainConfig{
		FaultProofStatus: script.FaultProofStatus{
			RespectedGameType: 42,
		},
	})
	require.NoError(t, err)

	// Verify the chain list was updated in memory
	for _, chain := range syncer.ChainList {
		if chain.ChainID == testChainID {
			require.Equal(t, uint32(42), chain.FaultProofStatus.RespectedGameType)
		}
	}

	// Test with invalid chain ID format
	err = syncer.UpdateChainList("not-a-number", script.ChainConfig{})
	require.Error(t, err)
}

func TestCodegenSyncer_SyncSingleChain(t *testing.T) {
	tempDir := t.TempDir()
	chainCfgs := createTestChainConfigs(t)

	// Get a chainID to test with (first one from configs)
	var testChainID uint64
	for id := range chainCfgs {
		testChainID = id
		break
	}

	// Modify the chain cfg
	cfg := chainCfgs[testChainID]
	cfg.FaultProofStatus = script.FaultProofStatus{
		RespectedGameType: 42,
	}
	chainCfgs[testChainID] = cfg

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelError, false))

	syncer, err := NewCodegenSyncer(lgr, "testdata", chainCfgs, WithOutputDirectory(tempDir))
	require.NoError(t, err)

	err = syncer.SyncSingleChain(fmt.Sprintf("%d", testChainID))
	require.NoError(t, err)

	// Verify chainList.json was created in tempDir and updated only for the specific chain
	chainListData, err := os.ReadFile(filepath.Join(tempDir, "chainList.json"))
	require.NoError(t, err)
	var chainList []config.ChainListEntry
	err = json.Unmarshal(chainListData, &chainList)
	require.NoError(t, err)

	for _, chain := range chainList {
		if chain.ChainID == testChainID {
			require.Equal(t, uint32(42), chain.FaultProofStatus.RespectedGameType)
		} else {
			require.Equal(t, uint32(0), chain.FaultProofStatus.RespectedGameType)
		}
	}

	// Verify chainList.toml was created
	_, err = os.Stat(filepath.Join(tempDir, "chainList.toml"))
	require.NoError(t, err)

	// Test with non-existent chain ID
	err = syncer.SyncSingleChain("999999")
	require.Error(t, err)
}

func TestCodegenSyncer_SyncAll(t *testing.T) {
	tempDir := t.TempDir()
	chainCfgs := createTestChainConfigs(t)

	for chainID := range chainCfgs {
		config := chainCfgs[chainID]
		config.FaultProofStatus = script.FaultProofStatus{
			RespectedGameType: 42,
		}
		chainCfgs[chainID] = config
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelError, false))

	syncer, err := NewCodegenSyncer(lgr, "testdata", chainCfgs, WithOutputDirectory(tempDir))
	require.NoError(t, err)

	err = syncer.SyncAll()
	require.NoError(t, err)

	// Verify chainList files were created in tempDir and updated for all chains
	chainListData, err := os.ReadFile(filepath.Join(tempDir, "chainList.json"))
	require.NoError(t, err)
	var chainList []config.ChainListEntry
	err = json.Unmarshal(chainListData, &chainList)
	require.NoError(t, err)

	for chainID := range chainCfgs {
		foundUpdatedChain := false
		for _, chain := range chainList {
			if chain.ChainID == chainID {
				require.Equal(t, uint32(42), chain.FaultProofStatus.RespectedGameType)
				foundUpdatedChain = true
			}
		}
		require.True(t, foundUpdatedChain, "Chain %d not found in written chainList.json", chainID)
	}

	// Verify chainList.toml was created and contains updated status
	chainListTomlData, err := os.ReadFile(filepath.Join(tempDir, "chainList.toml"))
	require.NoError(t, err)
	var chainListToml struct {
		Chains []config.ChainListEntry `toml:"chains"`
	}
	err = toml.Unmarshal(chainListTomlData, &chainListToml)
	require.NoError(t, err)

	for chainID := range chainCfgs {
		foundUpdatedChain := false
		for _, chain := range chainListToml.Chains {
			if chain.ChainID == chainID {
				require.Equal(t, uint32(42), chain.FaultProofStatus.RespectedGameType)
				foundUpdatedChain = true
			}
		}
		require.True(t, foundUpdatedChain, "Chain %d not found in written chainList.toml", chainID)
	}

	_, err = os.Stat(filepath.Join(tempDir, "CHAINS.md"))
	require.NoError(t, err)
}
