package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch/script"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum/go-ethereum/log"
)

// CodegenSyncer manages syncing of codegen files with on-chain data
type CodegenSyncer struct {
	lgr           log.Logger
	ChainList     []config.ChainListEntry
	Addresses     config.AddressesJSON
	addressesFile string
	chainListFile string
	inputDir      string
}

// Sync performs the complete sync process
func (s *CodegenSyncer) Sync() error {
	chainFiles, err := s.FindChainFiles()
	if err != nil {
		return err
	}

	if err := s.ReadAllChainFiles(chainFiles); err != nil {
		return err
	}

	return s.WriteFiles()
}

// NewCodegenSyncer creates a new syncer instance with provided file paths
func NewCodegenSyncer(lgr log.Logger, addressesFile, chainListFile, inputDir string) (*CodegenSyncer, error) {
	// Load addresses.json
	var addresses config.AddressesJSON
	addressesData, err := os.ReadFile(addressesFile)
	if err != nil {
		return nil, fmt.Errorf("error reading addresses file: %w", err)
	}
	if err := json.Unmarshal(addressesData, &addresses); err != nil {
		return nil, fmt.Errorf("error unmarshaling addresses.json: %w", err)
	}

	// Load chainList.json
	var chainList []config.ChainListEntry
	chainListData, err := os.ReadFile(chainListFile)
	if err != nil {
		return nil, fmt.Errorf("error reading chainList file: %w", err)
	}
	if err := json.Unmarshal(chainListData, &chainList); err != nil {
		return nil, fmt.Errorf("error unmarshaling chainList file: %w", err)
	}

	return &CodegenSyncer{
		lgr:           lgr,
		ChainList:     chainList,
		Addresses:     addresses,
		addressesFile: addressesFile,
		chainListFile: chainListFile,
		inputDir:      inputDir,
	}, nil
}

// FindChainFiles discovers and validates chain files in the input directory
func (s *CodegenSyncer) FindChainFiles() ([]string, error) {
	chainFiles, err := filepath.Glob(filepath.Join(s.inputDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("error finding chain files: %w", err)
	}
	if len(chainFiles) == 0 {
		return nil, fmt.Errorf("no chain files found in %s", s.inputDir)
	}
	s.lgr.Info("found fetcher input chain files", "count", len(chainFiles), "dir", s.inputDir)
	return chainFiles, nil
}

type chainConfig struct {
	Addresses        config.Addresses        `json:"addresses"`
	Roles            config.Roles            `json:"roles"`
	FaultProofStatus script.FaultProofStatus `json:"fault_proofs"`
}

// ReadChainFile parses a chain file into a struct
func (s *CodegenSyncer) ReadChainFile(file string) (*chainConfig, error) {
	chainData, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", file, err)
	}

	var chainConfig chainConfig
	if err := json.Unmarshal(chainData, &chainConfig); err != nil {
		return nil, fmt.Errorf("error unmarshaling chain file %s: %w", file, err)
	}

	return &chainConfig, nil
}

// ReadAllChainFiles updates addresses and chain list with data from chain files
func (s *CodegenSyncer) ReadAllChainFiles(chainFiles []string) error {
	for _, file := range chainFiles {
		baseFileName := filepath.Base(file)
		chainID := strings.TrimSuffix(baseFileName, filepath.Ext(baseFileName))
		s.lgr.Info("processing chain", "chainId", chainID)

		// Read and parse chain file
		chainConfig, err := s.ReadChainFile(file)
		if err != nil {
			return err
		}

		// Update addresses
		s.Addresses[chainID] = &config.AddressesWithRoles{
			Addresses: chainConfig.Addresses,
			Roles:     chainConfig.Roles,
		}

		// Update chainList
		if err := s.UpdateChainList(chainID, chainConfig.FaultProofStatus); err != nil {
			return err
		}
	}
	return nil
}

// UpdateChainList updates the ChainList entry for the given chain ID
func (s *CodegenSyncer) UpdateChainList(chainID string, faultProofStatus script.FaultProofStatus) error {
	chainIdUint64, err := strconv.ParseUint(chainID, 10, 64)
	if err != nil {
		return fmt.Errorf("error converting chainID to uint64: %w", err)
	}

	for i, entry := range s.ChainList {
		if entry.ChainID == chainIdUint64 {
			s.ChainList[i].FaultProofStatus = faultProofStatus
			break
		}
	}
	return nil
}

// WriteFiles writes all updated files to disk
func (s *CodegenSyncer) WriteFiles() error {
	// Write addresses.json
	updatedAddressesData, err := json.MarshalIndent(s.Addresses, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling updated addresses: %w", err)
	}
	if err := os.WriteFile(s.addressesFile, updatedAddressesData, 0o644); err != nil {
		return fmt.Errorf("error writing updated addresses.json: %w", err)
	}
	s.lgr.Info("successfully updated addresses.json", "chainCount", len(s.Addresses), "file", s.addressesFile)

	// Write chainList.json
	updatedChainListData, err := json.MarshalIndent(s.ChainList, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling updated chainList: %w", err)
	}
	if err := os.WriteFile(s.chainListFile, updatedChainListData, 0o644); err != nil {
		return fmt.Errorf("error writing updated chainList.json: %w", err)
	}
	s.lgr.Info("successfully updated chainList.json", "chainCount", len(s.ChainList), "file", s.chainListFile)

	// Write chainList.toml
	tomlFile := filepath.Join(filepath.Dir(s.chainListFile), "chainList.toml")

	// Create a wrapper struct for the TOML format
	chainListToml := struct {
		Chains []config.ChainListEntry `toml:"chains"`
	}{
		Chains: s.ChainList,
	}

	var buf strings.Builder
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(chainListToml); err != nil {
		return fmt.Errorf("error marshaling updated chainList to TOML: %w", err)
	}

	if err := os.WriteFile(tomlFile, []byte(buf.String()), 0o644); err != nil {
		return fmt.Errorf("error writing updated chainList.toml: %w", err)
	}
	s.lgr.Info("successfully updated chainList.toml", "chainCount", len(s.ChainList), "file", tomlFile)

	return nil
}
