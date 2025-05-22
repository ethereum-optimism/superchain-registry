package manage

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch/script"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/fs"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/log"
)

// CodegenSyncer manages syncing of codegen files with on-chain data
type CodegenSyncer struct {
	lgr         log.Logger
	ChainList   []config.ChainListEntry
	Addresses   config.AddressesJSON
	inputWd     string
	outputWd    string
	onchainCfgs map[uint64]script.ChainConfig
	diskCfgs    map[uint64]DiskChainConfig
}

type CodegenSyncerOption func(*CodegenSyncer)

func WithOutputDirectory(outputDir string) CodegenSyncerOption {
	return func(s *CodegenSyncer) {
		s.outputWd = outputDir
	}
}

func NewCodegenSyncer(lgr log.Logger, wd string, chainCfgs map[uint64]script.ChainConfig, opts ...CodegenSyncerOption) (*CodegenSyncer, error) {
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
	chainListData, err := os.ReadFile(paths.ChainListJsonFile(wd))
	if err != nil {
		return nil, fmt.Errorf("error reading chainList file: %w", err)
	}
	if err := json.Unmarshal(chainListData, &chainList); err != nil {
		return nil, fmt.Errorf("error unmarshaling chainList file: %w", err)
	}

	// Load disk chain configs
	diskChainCfgsSlice, err := CollectChainConfigs(paths.SuperchainConfigsDir(wd))
	lgr.Info("collected chain configs from disk", "numDiskCfgs", len(diskChainCfgsSlice))
	if err != nil {
		return nil, fmt.Errorf("error collecting chain configs: %w", err)
	}
	diskChainCfgs := make(map[uint64]DiskChainConfig)
	for _, cfg := range diskChainCfgsSlice {
		diskChainCfgs[cfg.Config.ChainID] = cfg
	}

	syncer := &CodegenSyncer{
		lgr:         lgr,
		ChainList:   chainList,
		Addresses:   addresses,
		inputWd:     wd,
		outputWd:    wd,
		onchainCfgs: chainCfgs,
		diskCfgs:    diskChainCfgs,
	}

	for _, opt := range opts {
		opt(syncer)
	}

	return syncer, nil
}

// SyncAll syncs codegen files with all entries in syncer.onchainCfgs
func (s *CodegenSyncer) SyncAll() error {
	if err := s.ProcessAllChains(); err != nil {
		return err
	}

	return s.WriteFiles()
}

// ProcessSingleChain updates syncer's internal data for a given chain
func (s *CodegenSyncer) ProcessSingleChain(chainId uint64, onchainCfg script.ChainConfig) error {
	s.lgr.Info("processing chain", "chainId", chainId)
	chainIdStr := strconv.FormatUint(chainId, 10)
	addressesWithRoles := config.CreateAddressesWithRolesFromFetcher(onchainCfg.Addresses, onchainCfg.Roles)
	s.Addresses[chainIdStr] = &addressesWithRoles

	if err := s.UpdateChainList(chainIdStr, onchainCfg); err != nil {
		return err
	}

	s.lgr.Info("finished processing chain", "chainId", chainId)
	return nil
}

// ProcessAllChains reads all input chain files and updates syncer's internal data accordingly
func (s *CodegenSyncer) ProcessAllChains() error {
	for chainId, cfg := range s.onchainCfgs {
		if err := s.ProcessSingleChain(chainId, cfg); err != nil {
			return err
		}
	}
	return nil
}

// UpdateChainList updates the ChainList entry for the given chain ID
func (s *CodegenSyncer) UpdateChainList(chainID string, onchainCfg script.ChainConfig) error {
	chainIdUint64, err := strconv.ParseUint(chainID, 10, 64)
	if err != nil {
		return fmt.Errorf("error converting chainID to uint64: %w", err)
	}

	diskCfg, ok := s.diskCfgs[chainIdUint64]
	if !ok {
		return fmt.Errorf("disk chain config not found for chain ID %s", chainID)
	}

	dir := filepath.Dir(diskCfg.Filepath)
	superchain := filepath.Base(dir)

	found := false
	chain := diskCfg.Config
	chainListEntry := chain.ChainListEntry(superchain, diskCfg.ShortName)

	if onchainCfg.FaultProofStatus == nil {
		chainListEntry.FaultProofs = config.FaultProofs{Status: "none"}
	} else if onchainCfg.FaultProofStatus.RespectedGameType == 1 {
		chainListEntry.FaultProofs = config.FaultProofs{Status: "permissioned"}
	} else {
		chainListEntry.FaultProofs = config.FaultProofs{Status: "permissionless"}
	}

	for i, entry := range s.ChainList {
		if entry.ChainID == chainIdUint64 {
			s.ChainList[i] = chainListEntry
			s.lgr.Info("updating existing chainList entry", "chainID", chainID)
			found = true
			break
		}
	}
	if !found {
		s.ChainList = append(s.ChainList, chainListEntry)
		s.lgr.Info("adding new chainList entry", "chainID", chainID)
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

	// Ensure the addresses directory exists
	addressesPath := paths.AddressesFile(s.outputWd)
	if err := paths.EnsureDir(filepath.Dir(addressesPath)); err != nil {
		return fmt.Errorf("error creating addresses.json directory: %w", err)
	}
	if err := os.WriteFile(addressesPath, updatedAddressesData, 0o644); err != nil {
		return fmt.Errorf("error writing updated addresses.json: %w", err)
	}
	s.lgr.Info("successfully updated addresses.json", "updatedChains", len(s.onchainCfgs), "totalChains", len(s.Addresses))

	// Write chainList.json
	updatedChainListData, err := json.MarshalIndent(s.ChainList, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling updated chainList: %w", err)
	}
	if err := os.WriteFile(paths.ChainListJsonFile(s.outputWd), updatedChainListData, 0o644); err != nil {
		return fmt.Errorf("error writing updated chainList.json: %w", err)
	}
	s.lgr.Info("successfully updated chainList.json", "updatedChains", len(s.onchainCfgs), "totalChains", len(s.ChainList))

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

	if err := os.WriteFile(paths.ChainListTomlFile(s.outputWd), []byte(buf.String()), 0o644); err != nil {
		return fmt.Errorf("error writing updated chainList.toml: %w", err)
	}
	s.lgr.Info("successfully updated chainList.toml", "updatedChains", len(s.onchainCfgs), "totalChains", len(s.ChainList))

	// Write CHAINS.md
	if err := GenChainsReadme(s.inputWd, paths.ChainMdFile(s.outputWd)); err != nil {
		return fmt.Errorf("error generating readme: %w", err)
	}
	s.lgr.Info("successfully updated CHAINS.md")

	return nil
}

//go:embed chains.md.tmpl
var chainsReadmeTemplateData string

var funcMap = template.FuncMap{
	"checkmark": func(in bool) string {
		if in {
			return "✅"
		}

		return "❌"
	},
	"optedInSuperchain": func(in *uint64) string {
		if in != nil && *in < uint64(time.Now().Unix()) {
			return "✅"
		}

		return "❌"
	},
}

var tmpl = template.Must(template.New("chains-readme").Funcs(funcMap).Parse(chainsReadmeTemplateData))

type ChainsReadmeData struct {
	Superchains []config.Superchain
	ChainData   [][]*config.Chain
}

func GenChainsReadme(rootP string, outP string) error {
	superchains, err := paths.Superchains(rootP)
	if err != nil {
		return fmt.Errorf("error getting superchains: %w", err)
	}

	var data ChainsReadmeData

	for _, superchain := range superchains {
		cfgs, err := CollectChainConfigs(paths.SuperchainDir(rootP, superchain))
		if err != nil {
			return fmt.Errorf("error collecting chain configs: %w", err)
		}

		chainData := make([]*config.Chain, len(cfgs))
		for i, cfg := range cfgs {
			chainData[i] = cfg.Config
		}

		data.Superchains = append(data.Superchains, superchain)
		data.ChainData = append(data.ChainData, chainData)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := fs.AtomicWrite(outP, 0o755, buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write chains readme: %w", err)
	}

	return nil
}
