package manage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/optimism/op-chain-ops/addresses"
	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch/script"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/common"
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
	data, err := os.ReadFile(paths.ChainListJsonFile("testdata"))
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
			chainCfgs[chainID] = convertToScriptChainConfig(t, chainAddrs, entry.FaultProofs)
		}
	}

	return chainCfgs
}

// convertToScriptChainConfig converts from config types to script types
func convertToScriptChainConfig(t *testing.T, chainAddrs *config.AddressesWithRoles, faultProofStatus config.FaultProofs) script.ChainConfig {
	var scriptAddrs script.Addresses
	var scriptRoles addresses.OpChainRoles

	// First, convert to a map of string addresses
	addressMap := make(map[string]string)
	bytes, err := json.Marshal(chainAddrs)
	require.NoError(t, err)
	err = json.Unmarshal(bytes, &addressMap)
	require.NoError(t, err)

	// Now populate the script structs manually using the map
	// For each field in scriptAddrs struct
	addressVal := reflect.ValueOf(&scriptAddrs).Elem()
	for i := 0; i < addressVal.NumField(); i++ {
		field := addressVal.Type().Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag == "" {
			jsonTag = field.Name
		}

		if addrStr, ok := addressMap[jsonTag]; ok && addrStr != "" {
			addr := common.HexToAddress(addrStr)
			addressVal.Field(i).Set(reflect.ValueOf(addr))
		}
	}

	// For each field in scriptRoles struct
	rolesVal := reflect.ValueOf(&scriptRoles).Elem()
	for i := 0; i < rolesVal.NumField(); i++ {
		field := rolesVal.Type().Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag == "" {
			jsonTag = field.Name
		}

		if addrStr, ok := addressMap[jsonTag]; ok && addrStr != "" {
			addr := common.HexToAddress(addrStr)
			rolesVal.Field(i).Set(reflect.ValueOf(addr))
		}
	}

	cfg := script.ChainConfig{
		Addresses: scriptAddrs,
		Roles:     scriptRoles,
	}
	if faultProofStatus.Status == "permissioned" {
		cfg.FaultProofStatus = &script.FaultProofStatus{
			Permissioned:      true,
			RespectedGameType: 1,
		}
	} else if faultProofStatus.Status == "permissionless" {
		cfg.FaultProofStatus = &script.FaultProofStatus{
			Permissioned:      true,
			Permissionless:    true,
			RespectedGameType: 0,
		}
	} else {
		cfg.FaultProofStatus = nil
	}
	return cfg
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
		FaultProofStatus: &script.FaultProofStatus{
			RespectedGameType: 0,
		},
	})
	require.NoError(t, err)

	// Verify the chain list was updated in memory
	for _, chain := range syncer.ChainList {
		if chain.ChainID == testChainID {
			require.Equal(t, "permissionless", chain.FaultProofs.Status)
		}
	}

	// Test with invalid chain ID format
	err = syncer.UpdateChainList("not-a-number", script.ChainConfig{})
	require.Error(t, err)
}

func TestCodegenSyncer_UpdateChainListSuperGameTypes(t *testing.T) {
	chainCfgs := createTestChainConfigs(t)
	var testChainID uint64
	for id := range chainCfgs {
		testChainID = id
		break
	}

	for _, test := range []struct {
		name              string
		respectedGameType uint32
		expectedStatus    string
	}{
		{name: "super permissioned", respectedGameType: 5, expectedStatus: "permissioned"},
		{name: "super cannon kona", respectedGameType: 9, expectedStatus: "permissionless"},
	} {
		t.Run(test.name, func(t *testing.T) {
			lgr := log.NewLogger(log.DiscardHandler())
			syncer, err := NewCodegenSyncer(lgr, "testdata", chainCfgs)
			require.NoError(t, err)

			err = syncer.UpdateChainList(fmt.Sprintf("%d", testChainID), script.ChainConfig{
				FaultProofStatus: &script.FaultProofStatus{
					RespectedGameType: test.respectedGameType,
				},
			})
			require.NoError(t, err)

			for _, chain := range syncer.ChainList {
				if chain.ChainID == testChainID {
					require.Equal(t, test.expectedStatus, chain.FaultProofs.Status)
					return
				}
			}
			require.Fail(t, "test chain not found")
		})
	}
}

func TestCodegenSyncer_SyncAll(t *testing.T) {
	tempDir := t.TempDir()
	chainCfgs := createTestChainConfigs(t)

	for chainID := range chainCfgs {
		config := chainCfgs[chainID]
		config.FaultProofStatus = &script.FaultProofStatus{
			RespectedGameType: 1,
		}
		chainCfgs[chainID] = config
	}

	lgr := log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelError, false))

	syncer, err := NewCodegenSyncer(lgr, "testdata", chainCfgs, WithOutputDirectory(tempDir))
	require.NoError(t, err)

	err = syncer.SyncAll()
	require.NoError(t, err)

	// Verify chainList files were created in tempDir and updated for all chains
	chainListData, err := os.ReadFile(paths.ChainListJsonFile(tempDir))
	require.NoError(t, err)
	var chainList []config.ChainListEntry
	err = json.Unmarshal(chainListData, &chainList)
	require.NoError(t, err)

	for chainID := range chainCfgs {
		foundUpdatedChain := false
		for _, chain := range chainList {
			if chain.ChainID == chainID {
				require.Equal(t, "permissioned", chain.FaultProofs.Status)
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
				require.Equal(t, "permissioned", chain.FaultProofs.Status)
				foundUpdatedChain = true
			}
		}
		require.True(t, foundUpdatedChain, "Chain %d not found in written chainList.toml", chainID)
	}

	_, err = os.Stat(filepath.Join(tempDir, "CHAINS.md"))
	require.NoError(t, err)
}

func TestCodegenSyncer_SyncAllUpdatesExistingAndPrunesRemovedChain(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	copyFile := func(src, dst string) {
		t.Helper()
		data, err := os.ReadFile(src)
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(filepath.Dir(dst), 0o755))
		require.NoError(t, os.WriteFile(dst, data, 0o644))
	}

	for _, filename := range []string{"superchain.toml", "op.toml", "testchain.toml"} {
		copyFile(
			filepath.Join("testdata", "superchain", "configs", "sepolia", filename),
			filepath.Join(inputDir, "superchain", "configs", "sepolia", filename),
		)
	}

	const removedChainID = uint64(999)
	chainList := loadTestChainList(t)
	chainList = append(chainList, config.ChainListEntry{ChainID: removedChainID})
	chainListData, err := json.Marshal(chainList)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(paths.ChainListJsonFile(inputDir), chainListData, 0o644))

	addresses := loadTestAddressesJSON(t)
	addresses[strconv.FormatUint(removedChainID, 10)] = &config.AddressesWithRoles{}
	addressesData, err := json.Marshal(addresses)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(paths.AddressesFile(inputDir)), 0o755))
	require.NoError(t, os.WriteFile(paths.AddressesFile(inputDir), addressesData, 0o644))

	chainCfgs := createTestChainConfigs(t)
	var updatedChainID uint64
	for chainID, chainCfg := range chainCfgs {
		updatedChainID = chainID
		chainCfg.FaultProofStatus = &script.FaultProofStatus{RespectedGameType: 0}
		chainCfgs = map[uint64]script.ChainConfig{chainID: chainCfg}
		break
	}

	lgr := log.NewLogger(log.DiscardHandler())
	syncer, err := NewCodegenSyncer(lgr, inputDir, chainCfgs, WithOutputDirectory(outputDir))
	require.NoError(t, err)
	require.NoError(t, syncer.SyncAll())

	data, err := os.ReadFile(paths.ChainListJsonFile(outputDir))
	require.NoError(t, err)
	var gotChainList []config.ChainListEntry
	require.NoError(t, json.Unmarshal(data, &gotChainList))

	foundUpdatedChain := false
	for _, entry := range gotChainList {
		require.NotEqual(t, removedChainID, entry.ChainID)
		if entry.ChainID == updatedChainID {
			require.Equal(t, "permissionless", entry.FaultProofs.Status)
			foundUpdatedChain = true
		}
	}
	require.True(t, foundUpdatedChain)

	data, err = os.ReadFile(paths.AddressesFile(outputDir))
	require.NoError(t, err)
	var gotAddresses config.AddressesJSON
	require.NoError(t, json.Unmarshal(data, &gotAddresses))
	require.NotContains(t, gotAddresses, strconv.FormatUint(removedChainID, 10))
}
