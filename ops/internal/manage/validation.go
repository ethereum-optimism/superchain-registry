package manage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BurntSushi/toml"
	opgenesis "github.com/ethereum-optimism/optimism/op-core/genesis"
	opparams "github.com/ethereum-optimism/optimism/op-core/params"
	"github.com/ethereum-optimism/optimism/op-core/superchain"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
)

const chainListUrl = "https://chainid.network/chains_mini.json"

var (
	ErrDuplicateChainID    = fmt.Errorf("duplicate chain ID")
	ErrDuplicateShortName  = fmt.Errorf("duplicate short name")
	ErrGenesisHashMismatch = fmt.Errorf("genesis hash mismatch")
)

type GlobalChainIDs struct {
	ChainIDs   map[uint64]bool
	ShortNames map[string]bool
}

func ValidateUniqueness(
	in *config.StagedChain,
	chains []DiskChainConfig,
) error {
	for _, chain := range chains {
		if chain.Config.ChainID == in.ChainID {
			return fmt.Errorf("%w: chains %s and %s: %d", ErrDuplicateChainID, chain.ShortName, in.ShortName, in.ChainID)
		}

		if chain.ShortName == in.ShortName {
			return fmt.Errorf("%w: chains %d and %d: %s", ErrDuplicateShortName, chain.Config.ChainID, in.ChainID, chain.ShortName)
		}
	}
	return nil
}

type ChainEntry struct {
	ChainID   uint64 `json:"chainId"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

func FetchGlobalChainIDs() (map[uint64]ChainEntry, error) {
	req, err := http.NewRequest(http.MethodGet, chainListUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "optimism-superchain-registry-validation")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var entries []ChainEntry
	if err := json.NewDecoder(res.Body).Decode(&entries); err != nil {
		return nil, err
	}

	out := make(map[uint64]ChainEntry)
	for _, entry := range entries {
		out[entry.ChainID] = entry
	}
	return out, nil
}

// ValidateGenesisIntegrity checks that the genesis JSON, with a chain config derived from cfg,
// hashes to the genesis block hash in cfg.
func ValidateGenesisIntegrity(cfg *config.Chain, genesisJSON json.RawMessage) error {
	var genesis opgenesis.Genesis
	if err := json.Unmarshal(genesisJSON, &genesis); err != nil {
		return fmt.Errorf("failed to decode genesis: %w", err)
	}
	chainConfig, err := opChainConfig(cfg)
	if err != nil {
		return err
	}
	genesis.Config = chainConfig

	block := genesis.ToBlock()
	if block.Hash() != cfg.Genesis.L2.Hash {
		return fmt.Errorf("%w: expected %s, got %s", ErrGenesisHashMismatch, cfg.Genesis.L2.Hash.Hex(), block.Hash().Hex())
	}

	return nil
}

// opChainConfig returns the chain config of cfg's chain, as op-core builds it from the registry's
// chain config. It fails for a hardfork field that op-core's registry types do not know.
func opChainConfig(cfg *config.Chain) (*opparams.ChainConfig, error) {
	hardforks, err := toml.Marshal(cfg.Hardforks)
	if err != nil {
		return nil, fmt.Errorf("failed to encode hardforks: %w", err)
	}
	chain := &superchain.ChainConfig{
		ChainID: cfg.ChainID,
		Optimism: &superchain.OptimismConfig{
			EIP1559Elasticity:  cfg.Optimism.EIP1559Elasticity,
			EIP1559Denominator: cfg.Optimism.EIP1559Denominator,
		},
	}
	if err := jsonutil.DecodeTOMLStrict(bytes.NewReader(hardforks), &chain.Hardforks, nil); err != nil {
		return nil, fmt.Errorf("failed to decode hardforks with op-core: %w", err)
	}
	if cfg.Optimism.EIP1559DenominatorCanyon != 0 {
		chain.Optimism.EIP1559DenominatorCanyon = &cfg.Optimism.EIP1559DenominatorCanyon
	}
	return chain.OpChainConfig(), nil
}
