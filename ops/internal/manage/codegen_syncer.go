package manage

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch/script"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/log"
)

// CodegenSyncer manages syncing of codegen files with on-chain data
type CodegenSyncer struct {
	lgr       log.Logger
	ChainList []config.ChainListEntry
	Addresses config.AddressesJSON
	wd        string
	chainCfgs map[uint64]script.ChainConfig
}

// SyncSingleChain syncs only the specified chain
func (s *CodegenSyncer) SyncSingleChain(chainId string) error {
	chainIdUint64, err := strconv.ParseUint(chainId, 10, 64)
	if err != nil {
		return fmt.Errorf("error converting chainID to uint64: %w", err)
	}
	cfg, ok := s.chainCfgs[chainIdUint64]
	if !ok {
		return fmt.Errorf("chain config not found for chain ID %s", chainId)
	}
	s.lgr.Info("found chain config", "chainId", chainId)

	if err := s.ProcessSingleChain(chainIdUint64, cfg); err != nil {
		return err
	}

	return s.WriteFiles()
}

// SyncAll performs the complete sync process
func (s *CodegenSyncer) SyncAll() error {
	if err := s.ProcessAllChains(); err != nil {
		return err
	}

	return s.WriteFiles()
}

// NewCodegenSyncer creates a new syncer instance with provided file paths
func NewCodegenSyncer(lgr log.Logger, wd string, chainCfgs map[uint64]script.ChainConfig) (*CodegenSyncer, error) {
	// Load addresses.json data
	var addresses config.AddressesJSON
	addressesData, err := os.ReadFile(paths.AddressesFile(wd))
	if err != nil {
		return nil, fmt.Errorf("error reading addresses file: %w", err)
	}
	if err := json.Unmarshal(addressesData, &addresses); err != nil {
		return nil, fmt.Errorf("error unmarshaling addresses.json: %w", err)
	}

	// Load chainList.json data
	var chainList []config.ChainListEntry
	chainListData, err := os.ReadFile(filepath.Join(wd, "chainList.json"))
	if err != nil {
		return nil, fmt.Errorf("error reading chainList file: %w", err)
	}
	if err := json.Unmarshal(chainListData, &chainList); err != nil {
		return nil, fmt.Errorf("error unmarshaling chainList file: %w", err)
	}

	return &CodegenSyncer{
		lgr:       lgr,
		ChainList: chainList,
		Addresses: addresses,
		wd:        wd,
		chainCfgs: chainCfgs,
	}, nil
}

// ProcessSingleChain updates syncer's internal data for a given chain
func (s *CodegenSyncer) ProcessSingleChain(chainId uint64, cfg script.ChainConfig) error {
	chainIdStr := strconv.FormatUint(chainId, 10)
	s.Addresses[chainIdStr] = &config.AddressesWithRoles{
		Addresses: cfg.Addresses,
		Roles:     cfg.Roles,
	}

	if err := s.UpdateChainList(chainIdStr, cfg); err != nil {
		return err
	}

	return nil
}

// ProcessAllChains reads all input chain files and updates syncer's internal data accordingly
func (s *CodegenSyncer) ProcessAllChains() error {
	for chainId, cfg := range s.chainCfgs {
		s.lgr.Info("processing chain", "chainId", chainId)

		if err := s.ProcessSingleChain(chainId, cfg); err != nil {
			return err
		}
	}
	return nil
}

// UpdateChainList updates the ChainList entry for the given chain ID
func (s *CodegenSyncer) UpdateChainList(chainID string, cfg script.ChainConfig) error {
	chainIdUint64, err := strconv.ParseUint(chainID, 10, 64)
	if err != nil {
		return fmt.Errorf("error converting chainID to uint64: %w", err)
	}

	for i, entry := range s.ChainList {
		if entry.ChainID == chainIdUint64 {
			s.ChainList[i].FaultProofStatus = cfg.FaultProofStatus
			break
		}
	}
	return nil
}

// WriteFiles writes all updated data to disk
func (s *CodegenSyncer) WriteFiles() error {
	// Write addresses.json
	updatedAddressesData, err := json.MarshalIndent(s.Addresses, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling updated addresses: %w", err)
	}
	if err := os.WriteFile(paths.AddressesFile(s.wd), updatedAddressesData, 0o644); err != nil {
		return fmt.Errorf("error writing updated addresses.json: %w", err)
	}
	s.lgr.Info("successfully updated addresses.json", "chainCount", len(s.Addresses))

	// Write chainList.json
	updatedChainListData, err := json.MarshalIndent(s.ChainList, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling updated chainList: %w", err)
	}
	if err := os.WriteFile(filepath.Join(s.wd, "chainList.json"), updatedChainListData, 0o644); err != nil {
		return fmt.Errorf("error writing updated chainList.json: %w", err)
	}
	s.lgr.Info("successfully updated chainList.json", "chainCount", len(s.ChainList))

	// Write chainList.toml
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

	if err := os.WriteFile(filepath.Join(s.wd, "chainList.toml"), []byte(buf.String()), 0o644); err != nil {
		return fmt.Errorf("error writing updated chainList.toml: %w", err)
	}
	s.lgr.Info("successfully updated chainList.toml", "chainCount", len(s.ChainList))

	// Write CHAINS.md
	if err := GenChainsReadme(s.wd, path.Join(s.wd, "CHAINS.md")); err != nil {
		return fmt.Errorf("error generating readme: %w", err)
	}
	s.lgr.Info("successfully updated CHAINS.md")

	return nil
}
