package manage

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
)

var (
	errDepsetLengths       = errors.New("inconsistent depset lengths")
	errInconsistentDepsets = errors.New("inconsistent depset values")
	errMissingAddress      = errors.New("missing address")
	errMismatchedL1        = errors.New("depset members settle to different L1s")
	errMismatchedSCAddr    = errors.New("depset members use different SuperchainConfig addresses")
	errMissingSuperchain   = errors.New("missing superchain definition")
)

type DepsetChecker struct {
	lgr             log.Logger
	addrs           config.AddressesJSON
	superchains     map[config.Superchain]*config.SuperchainDefinition
	diskChainCfgs   map[uint64]DiskChainConfig
	processedChains map[uint64]bool
	chainsProcessed int
}

// NewDepsetChecker returns a checker for the depsets in cfgs. superchains holds the superchain.toml
// definition of every superchain dir in cfgs; it is only consulted for depsets whose members span
// more than one superchain dir.
func NewDepsetChecker(
	logger log.Logger,
	cfgs []DiskChainConfig,
	addrs config.AddressesJSON,
	superchains map[config.Superchain]*config.SuperchainDefinition,
) *DepsetChecker {
	dc := &DepsetChecker{
		lgr:             logger,
		addrs:           addrs,
		superchains:     superchains,
		diskChainCfgs:   make(map[uint64]DiskChainConfig),
		processedChains: make(map[uint64]bool),
	}

	for _, cfg := range cfgs {
		dc.diskChainCfgs[cfg.Config.ChainID] = cfg
	}

	logger.Info("loaded depset checker instance with chain configs", "numDiskCfgs", len(cfgs))
	return dc
}

func (dc *DepsetChecker) Check() error {
	for _, cfg := range dc.diskChainCfgs {
		if dc.processedChains[cfg.Config.ChainID] {
			// already processed this chain (it may be revisited during checkOffchain)
			continue
		}
		if cfg.Config.Interop == nil {
			// nothing to check for this chain
			dc.lgr.Info("skipping chain with no interop config", "chainID", cfg.Config.ChainID)
			dc.processedChains[cfg.Config.ChainID] = true
			dc.chainsProcessed++
			continue
		}

		// collect all chains in the depset, then process them
		depsetCfgs := []DiskChainConfig{}
		depset := cfg.Config.Interop.Dependencies

		for chainId := range depset {
			chainIdUint64, err := strconv.ParseUint(chainId, 10, 64)
			if err != nil {
				return fmt.Errorf("chain ID cannot be converted to a uint64: %s", chainId)
			}
			if _, ok := dc.diskChainCfgs[chainIdUint64]; !ok {
				return fmt.Errorf("chain ID %d not found in diskChainCfgs", chainIdUint64)
			}
			depsetCfgs = append(depsetCfgs, dc.diskChainCfgs[chainIdUint64])
		}

		if err := dc.checkOffchain(depsetCfgs); err != nil {
			return fmt.Errorf("invalid depset (offchain consistency): %w", err)
		}
		if err := dc.checkOnchain(depsetCfgs); err != nil {
			return fmt.Errorf("invalid depset (onchain addresses): %w", err)
		}

		// mark all chains in the depset as processed to avoid repeats
		for _, dep := range depsetCfgs {
			dc.lgr.Info("processed chain", "chainID", dep.Config.ChainID)
			dc.processedChains[dep.Config.ChainID] = true
			dc.chainsProcessed++
		}
	}

	numChains := len(dc.diskChainCfgs)
	if dc.chainsProcessed != numChains {
		return fmt.Errorf("chainsProcessed (%d) does not match totalChainCfgs (%d)", dc.chainsProcessed, numChains)
	}
	dc.lgr.Info("processed all chains", "numChains", numChains)

	return nil
}

// checkOffchain ensures that the depsets encoded in the chain configs are valid
// and consistent across all configs in the same set
func (dc *DepsetChecker) checkOffchain(cfgs []DiskChainConfig) error {
	if len(cfgs) == 0 {
		return fmt.Errorf("no chain configs provided to checkOffchain")
	}

	var firstDepset map[string]config.StaticConfigDependency
	for i, cfg := range cfgs {
		if i == 0 {
			firstDepset = cfg.Config.Interop.Dependencies
		}

		// check that all chains in the depset settle to the same L1 and share a SuperchainConfig
		if err := dc.checkSameSuperchainConfig(cfgs[0], cfg); err != nil {
			return err
		}

		// check equality of dependency sets
		thisDepset := cfg.Config.Interop.Dependencies
		if i > 0 {
			// 1. check if the maps have the same length
			if len(firstDepset) != len(thisDepset) {
				return fmt.Errorf("%w: chain %d has different number of dependencies than first chain (%d vs %d)",
					errDepsetLengths,
					cfg.Config.ChainID, len(thisDepset), len(firstDepset))
			}

			// 2. check that depset members are the same across all configs
			for chainID := range firstDepset {
				_, exists := thisDepset[chainID]
				if !exists {
					return fmt.Errorf("%w: chain %d is missing (transitive) dependency for chain ID %s (direct dependency of %d)",
						errInconsistentDepsets,
						cfg.Config.ChainID, chainID, cfgs[0].Config.ChainID)
				}
			}
		}
	}

	return nil
}

// checkSameSuperchainConfig ensures that two depset members settle to the same L1 and use the same
// SuperchainConfig. Members may live in different superchain dirs (e.g. one dir per devnet), so a
// depset never forces chains to move dirs, which would rename their node network names.
func (dc *DepsetChecker) checkSameSuperchainConfig(first, cfg DiskChainConfig) error {
	if cfg.Superchain == first.Superchain {
		return nil
	}

	firstDef, ok := dc.superchains[first.Superchain]
	if !ok || firstDef == nil {
		return fmt.Errorf("%w: %s (chain %d)", errMissingSuperchain, first.Superchain, first.Config.ChainID)
	}
	def, ok := dc.superchains[cfg.Superchain]
	if !ok || def == nil {
		return fmt.Errorf("%w: %s (chain %d)", errMissingSuperchain, cfg.Superchain, cfg.Config.ChainID)
	}

	if def.L1.ChainID != firstDef.L1.ChainID {
		return fmt.Errorf("%w: chain %d (superchain %s) settles to L1 %d, chain %d (superchain %s) settles to L1 %d",
			errMismatchedL1,
			cfg.Config.ChainID, cfg.Superchain, def.L1.ChainID,
			first.Config.ChainID, first.Superchain, firstDef.L1.ChainID)
	}

	zeroAddress := config.ChecksummedAddress{}
	if firstDef.SuperchainConfigAddr == nil || *firstDef.SuperchainConfigAddr == zeroAddress {
		return fmt.Errorf("%w: superchain %s has no superchain_config_addr", errMismatchedSCAddr, first.Superchain)
	}
	if def.SuperchainConfigAddr == nil || *def.SuperchainConfigAddr != *firstDef.SuperchainConfigAddr {
		return fmt.Errorf("%w: chain %d (superchain %s) uses %v, chain %d (superchain %s) uses %s",
			errMismatchedSCAddr,
			cfg.Config.ChainID, cfg.Superchain, def.SuperchainConfigAddr,
			first.Config.ChainID, first.Superchain, firstDef.SuperchainConfigAddr)
	}

	return nil
}

// checkOnchain ensures that DisputeGameFactoryProxy and EthLockboxProxy addresses read from onchain
// are the same for all chains in a depset. These checks are only performed for chains who have
// already activated interop (i.e. lagoon_time < current unix timestamp).
func (dc *DepsetChecker) checkOnchain(cfgs []DiskChainConfig) error {
	if len(cfgs) == 0 {
		return fmt.Errorf("no chain configs provided to checkOnchain")
	}

	if len(cfgs) == 1 {
		return nil
	}

	// Find chains that have already activated interop
	now := time.Now().Unix()
	var activatedChains []DiskChainConfig
	for _, cfg := range cfgs {
		lagoonTime := cfg.Config.Hardforks.LagoonTime
		if lagoonTime == nil || *lagoonTime.U64Ptr() > uint64(now) {
			continue
		}
		activatedChains = append(activatedChains, cfg)
	}

	// If we have fewer than 2 activated chains, no comparison needed
	if len(activatedChains) < 2 {
		dc.lgr.Info("skipping onchain checks for depset")
		return nil
	}

	// Get and validate the first valid chain's addresses
	firstAddrs, err := dc.getAndValidateAddresses(activatedChains[0].Config.ChainID)
	if err != nil {
		return err
	}

	firstDisputeGameFactoryProxy := strings.ToLower((*firstAddrs.DisputeGameFactoryProxy).String())
	firstEthLockboxProxy := strings.ToLower((*firstAddrs.EthLockboxProxy).String())

	// Check all remaining valid chains in the dependency set
	for i := 1; i < len(activatedChains); i++ {
		cfg := activatedChains[i]
		addrs, err := dc.getAndValidateAddresses(cfg.Config.ChainID)
		if err != nil {
			return err
		}

		if strings.ToLower((*addrs.DisputeGameFactoryProxy).String()) != firstDisputeGameFactoryProxy {
			return fmt.Errorf("DisputeGameFactoryProxy address mismatch for chain %d, expected %s, got %s",
				cfg.Config.ChainID, firstDisputeGameFactoryProxy, addrs.DisputeGameFactoryProxy)
		}
		if strings.ToLower((*addrs.EthLockboxProxy).String()) != firstEthLockboxProxy {
			return fmt.Errorf("EthLockboxProxy address mismatch for chain %d, expected %s, got %s",
				cfg.Config.ChainID, firstEthLockboxProxy, addrs.EthLockboxProxy)
		}
	}

	return nil
}

// getAndValidateAddresses retrieves and validates the proxy addresses for a given chain ID.
// Returns the addresses and an error if validation fails.
func (dc *DepsetChecker) getAndValidateAddresses(chainID uint64) (*config.AddressesWithRoles, error) {
	chainIdStr := eth.ChainIDFromUInt64(chainID).String()
	addrs, ok := dc.addrs[chainIdStr]
	if !ok {
		return nil, fmt.Errorf("%w: no addresses found for chain %d", errMissingAddress, chainID)
	}
	if addrs == nil {
		return nil, fmt.Errorf("%w: no addresses found for chain %d", errMissingAddress, chainID)
	}

	zeroAddress := config.ChecksummedAddress{}
	if addrs.DisputeGameFactoryProxy == nil || *addrs.DisputeGameFactoryProxy == zeroAddress {
		return nil, fmt.Errorf("%w: no DisputeGameFactoryProxy found for chain %d", errMissingAddress, chainID)
	}
	if addrs.EthLockboxProxy == nil || *addrs.EthLockboxProxy == zeroAddress {
		return nil, fmt.Errorf("%w: no EthLockboxProxy found for chain %d", errMissingAddress, chainID)
	}
	return addrs, nil
}

// LoadSuperchainDefinitions reads the superchain.toml of every superchain dir in the repo at wd,
// keyed by dir name.
func LoadSuperchainDefinitions(wd string) (map[config.Superchain]*config.SuperchainDefinition, error) {
	superchains, err := paths.Superchains(wd)
	if err != nil {
		return nil, fmt.Errorf("failed to list superchains: %w", err)
	}

	defs := make(map[config.Superchain]*config.SuperchainDefinition, len(superchains))
	for _, superchain := range superchains {
		var def config.SuperchainDefinition
		if err := paths.ReadTOMLFile(paths.SuperchainDefinitionPath(wd, superchain), &def); err != nil {
			return nil, fmt.Errorf("failed to read superchain definition for %s: %w", superchain, err)
		}
		defs[superchain] = &def
	}
	return defs, nil
}
