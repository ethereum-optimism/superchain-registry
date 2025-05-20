package manage

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum/go-ethereum/log"
)

var (
	errInvalidActivationTime = errors.New("activation_time is before interop_time")
	errDepsetLengths         = errors.New("inconsistent depset lengths")
	errInconsistentDepsets   = errors.New("inconsistent depset values")
	errMissingAddress        = errors.New("missing address")
	errDuplicateChainIndex   = errors.New("duplicate chain index")
)

type DepsetChecker struct {
	lgr             log.Logger
	addrs           config.AddressesJSON
	diskChainCfgs   map[uint64]DiskChainConfig
	processedChains map[uint64]bool
	chainsProcessed int
}

func NewDepsetChecker(logger log.Logger, cfgs []DiskChainConfig, addrs config.AddressesJSON) (*DepsetChecker, error) {
	dc := &DepsetChecker{
		lgr:             logger,
		addrs:           addrs,
		diskChainCfgs:   make(map[uint64]DiskChainConfig),
		processedChains: make(map[uint64]bool),
	}

	for _, cfg := range cfgs {
		dc.diskChainCfgs[cfg.Config.ChainID] = cfg
	}

	logger.Info("loaded depset checker instance with chain configs", "numDiskCfgs", len(cfgs))
	return dc, nil
}

func (dc *DepsetChecker) Check() error {
	for _, cfg := range dc.diskChainCfgs {
		if dc.processedChains[cfg.Config.ChainID] {
			// already processed this chain
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
			depsetCfgs = append(depsetCfgs, dc.diskChainCfgs[chainIdUint64])
		}

		if err := dc.checkOffchain(depsetCfgs); err != nil {
			return fmt.Errorf("invalid depset (offchain consistency): %w", err)
		}
		// TODO: re-enable this once we can pull in the updated op-fetcher from monorepo
		//if err := dc.checkOnchain(depsetCfgs); err != nil {
		//return fmt.Errorf("invalid depset (onchain addresses): %w", err)
		//}

		// mark all chains in the depset as processed to avoid repeats
		for _, dep := range depsetCfgs {
			dc.lgr.Info("processed chain", "chainID", dep.Config.ChainID)
			dc.processedChains[dep.Config.ChainID] = true
			dc.chainsProcessed++
		}
	}

	numChains := len(dc.diskChainCfgs)
	if dc.chainsProcessed != numChains {
		return fmt.Errorf("chainsProcessed (%d) does not match totalChainCgs (%d)", dc.chainsProcessed, numChains)
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

	var firstSuperchain string
	var firstDepset map[string]config.Dependency
	for i, cfg := range cfgs {
		if i == 0 {
			firstSuperchain = cfg.Superchain
			firstDepset = cfg.Config.Interop.Dependencies

			// Check for duplicate chain indices in the first config
			chainIndices := make(map[uint32]string, len(firstDepset))
			for chainID, dep := range firstDepset {
				if existingChainID, exists := chainIndices[dep.ChainIndex]; exists {
					return fmt.Errorf("%w: index %d used for chains %s and %s",
						errDuplicateChainIndex, dep.ChainIndex, existingChainID, chainID)
				}
				chainIndices[dep.ChainIndex] = chainID
			}

		}

		// check that all chains in the depset are part of the same superchain
		if cfg.Superchain != firstSuperchain {
			return fmt.Errorf("chain %d is part of superchain %s, expected superchain %s", cfg.Config.ChainID, cfg.Superchain, firstSuperchain)
		}

		// check that activation_time is valid
		thisDepset := cfg.Config.Interop.Dependencies
		chainId := strconv.FormatUint(cfg.Config.ChainID, 10)
		if cfg.Config.Hardforks.InteropTime == nil {
			return fmt.Errorf("chain %d has no hardfork timestamp for interop", cfg.Config.ChainID)
		}
		interopTime := *cfg.Config.Hardforks.InteropTime.U64Ptr()
		if thisDepset[chainId].ActivationTime < interopTime {
			return fmt.Errorf("%w: chainId %d, activation_time %d interop_time %d",
				errInvalidActivationTime,
				cfg.Config.ChainID,
				thisDepset[chainId].ActivationTime,
				interopTime)
		}

		// check deep equality of dependency sets
		if i > 0 {
			// 1. check if the maps have the same length
			if len(firstDepset) != len(thisDepset) {
				return fmt.Errorf("%w: chain %d has different number of dependencies than first chain (%d vs %d)",
					errDepsetLengths,
					cfg.Config.ChainID, len(thisDepset), len(firstDepset))
			}

			// 2. check for deep equality of dependency values
			for chainID, dep := range firstDepset {
				otherDep, exists := thisDepset[chainID]
				if !exists {
					return fmt.Errorf("%w: chain %d is missing dependency for chain ID %s that exists in first chain",
						errInconsistentDepsets,
						cfg.Config.ChainID, chainID)
				}

				if dep.ChainIndex != otherDep.ChainIndex {
					return fmt.Errorf("%w: chain %d has different chain index for dependency %s (%d vs %d)",
						errInconsistentDepsets,
						cfg.Config.ChainID, chainID, otherDep.ChainIndex, dep.ChainIndex)
				}
				if dep.ActivationTime != otherDep.ActivationTime {
					return fmt.Errorf("%w: chain %d has different activation time for dependency %s (%d vs %d)",
						errInconsistentDepsets,
						cfg.Config.ChainID, chainID, otherDep.ActivationTime, dep.ActivationTime)
				}
			}
		}
	}

	return nil
}

// checkOnchain ensures that DisputeGameFactoryProxy and EthLockboxProxy addresses read from onchain
// are the same for all chains in a depset
func (dc *DepsetChecker) checkOnchain(cfgs []DiskChainConfig) error {
	if len(cfgs) == 0 {
		return fmt.Errorf("no chain configs provided to checkOnchain")
	}

	if len(cfgs) == 1 {
		return nil
	}

	// Get and validate the first chain's addresses
	firstAddrs, err := dc.getAndValidateAddresses(cfgs[0].Config.ChainID)
	if err != nil {
		return err
	}

	firstDisputeGameFactoryProxy := strings.ToLower((*firstAddrs.DisputeGameFactoryProxy).String())
	firstEthLockboxProxy := strings.ToLower((*firstAddrs.EthLockboxProxy).String())

	// Check all chains in the dependency set
	for _, cfg := range cfgs {
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
