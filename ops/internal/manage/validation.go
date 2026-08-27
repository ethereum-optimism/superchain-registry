package manage

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"
)

const chainListUrl = "https://chainid.network/chains_mini.json"

var (
	ErrDuplicateChainID            = fmt.Errorf("duplicate chain ID")
	ErrDuplicateShortName          = fmt.Errorf("duplicate short name")
	ErrGenesisHashMismatch         = fmt.Errorf("genesis hash mismatch")
	ErrUnsupportedDataAvailability = fmt.Errorf("unsupported data availability for new registry entry")
)

// legacyAltDAChains identifies the existing deployed networks whose registry
// metadata must remain readable. No new entries may be added to this list.
var legacyAltDAChains = map[string]uint64{
	"mainnet/automata":            65536,
	"mainnet/celo":                42220,
	"mainnet/cyber":               7560,
	"mainnet/fraxtal":             252,
	"mainnet/funki":               33979,
	"mainnet/lyra":                957,
	"mainnet/orderly":             291,
	"mainnet/redstone":            690,
	"mainnet/silent-data-mainnet": 380929,
	"mainnet/xterio-eth":          2702128,
	"sepolia/celo-sep":            11142220,
	"sepolia/funki":               3397901,
}

type GlobalChainIDs struct {
	ChainIDs   map[uint64]bool
	ShortNames map[string]bool
}

// ValidateNewChainDataAvailability restricts registry intake to Ethereum DA without changing existing chain records.
func ValidateNewChainDataAvailability(cfg *config.Chain) error {
	if cfg.DataAvailabilityType != "eth-da" {
		return fmt.Errorf(
			"%w: data_availability_type must be %q, got %q",
			ErrUnsupportedDataAvailability,
			"eth-da",
			cfg.DataAvailabilityType,
		)
	}
	if cfg.AltDA != nil {
		return fmt.Errorf("%w: alt_da must not be configured", ErrUnsupportedDataAvailability)
	}
	if cfg.Addresses.DAChallengeAddress != nil {
		return fmt.Errorf("%w: addresses.DAChallengeAddress must not be configured", ErrUnsupportedDataAvailability)
	}
	return nil
}

// ValidateRegistryDataAvailability enforces Ethereum DA for the complete
// registry tree while preserving the exact set of existing legacy networks.
func ValidateRegistryDataAvailability(chains []DiskChainConfig) error {
	seenLegacy := make(map[string]bool, len(legacyAltDAChains))
	for _, chain := range chains {
		key := fmt.Sprintf("%s/%s", chain.Superchain, chain.ShortName)
		legacyChainID, isLegacy := legacyAltDAChains[key]
		if !isLegacy {
			if err := ValidateNewChainDataAvailability(chain.Config); err != nil {
				return fmt.Errorf("chain %s: %w", key, err)
			}
			continue
		}

		if chain.Config.ChainID != legacyChainID {
			return fmt.Errorf("legacy Alt-DA chain %s must have chain ID %d, got %d", key, legacyChainID, chain.Config.ChainID)
		}
		if chain.Config.DataAvailabilityType != "alt-da" {
			return fmt.Errorf("legacy Alt-DA chain %s must retain data_availability_type %q", key, "alt-da")
		}
		if chain.Config.Addresses.DAChallengeAddress != nil {
			return fmt.Errorf("legacy Alt-DA chain %s contains retired addresses.DAChallengeAddress", key)
		}
		if seenLegacy[key] {
			return fmt.Errorf("legacy Alt-DA chain %s appears more than once", key)
		}
		seenLegacy[key] = true
	}

	for key := range legacyAltDAChains {
		if !seenLegacy[key] {
			return fmt.Errorf("legacy Alt-DA chain %s is missing from the registry", key)
		}
	}
	return nil
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

func ValidateGenesisIntegrity(cfg *config.Chain, genesis *core.Genesis) error {
	genesisActivation := uint64(0)
	out := &params.ChainConfig{
		ChainID:                 new(big.Int).SetUint64(cfg.ChainID),
		HomesteadBlock:          common.Big0,
		DAOForkBlock:            nil,
		DAOForkSupport:          false,
		EIP150Block:             common.Big0,
		EIP155Block:             common.Big0,
		EIP158Block:             common.Big0,
		ByzantiumBlock:          common.Big0,
		ConstantinopleBlock:     common.Big0,
		PetersburgBlock:         common.Big0,
		IstanbulBlock:           common.Big0,
		MuirGlacierBlock:        common.Big0,
		BerlinBlock:             common.Big0,
		LondonBlock:             common.Big0,
		ArrowGlacierBlock:       common.Big0,
		GrayGlacierBlock:        common.Big0,
		MergeNetsplitBlock:      common.Big0,
		ShanghaiTime:            cfg.Hardforks.CanyonTime.U64Ptr(),  // Shanghai activates with Canyon
		CancunTime:              cfg.Hardforks.EcotoneTime.U64Ptr(), // Cancun activates with Ecotone
		PragueTime:              cfg.Hardforks.IsthmusTime.U64Ptr(), // Prague activates with Isthmus
		BedrockBlock:            common.Big0,
		RegolithTime:            &genesisActivation,
		CanyonTime:              cfg.Hardforks.CanyonTime.U64Ptr(),
		EcotoneTime:             cfg.Hardforks.EcotoneTime.U64Ptr(),
		FjordTime:               cfg.Hardforks.FjordTime.U64Ptr(),
		GraniteTime:             cfg.Hardforks.GraniteTime.U64Ptr(),
		HoloceneTime:            cfg.Hardforks.HoloceneTime.U64Ptr(),
		IsthmusTime:             cfg.Hardforks.IsthmusTime.U64Ptr(),
		LagoonTime:              cfg.Hardforks.LagoonTime.U64Ptr(),
		JovianTime:              cfg.Hardforks.JovianTime.U64Ptr(),
		TerminalTotalDifficulty: common.Big0,
		Ethash:                  nil,
		Clique:                  nil,
	}

	out.Optimism = &params.OptimismConfig{
		EIP1559Elasticity:  cfg.Optimism.EIP1559Elasticity,
		EIP1559Denominator: cfg.Optimism.EIP1559Denominator,
	}

	if cfg.Optimism.EIP1559DenominatorCanyon != 0 {
		out.Optimism.EIP1559DenominatorCanyon = &cfg.Optimism.EIP1559DenominatorCanyon
	}

	genCopy := &core.Genesis{
		Config:        out,
		Nonce:         genesis.Nonce,
		Timestamp:     genesis.Timestamp,
		ExtraData:     genesis.ExtraData,
		GasLimit:      genesis.GasLimit,
		Difficulty:    genesis.Difficulty,
		Mixhash:       genesis.Mixhash,
		Coinbase:      genesis.Coinbase,
		Alloc:         genesis.Alloc,
		Number:        genesis.Number,
		GasUsed:       genesis.GasUsed,
		ParentHash:    genesis.ParentHash,
		BaseFee:       genesis.BaseFee,
		ExcessBlobGas: genesis.ExcessBlobGas,
		BlobGasUsed:   genesis.BlobGasUsed,
		StateHash:     genesis.StateHash,
	}
	block := genCopy.ToBlock()

	if block.Hash() != cfg.Genesis.L2.Hash {
		return fmt.Errorf("%w: expected %s, got %s", ErrGenesisHashMismatch, cfg.Genesis.L2.Hash.Hex(), block.Hash().Hex())
	}

	return nil
}
