package manage

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/report"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	errReviewPeriodOutOfRange = errors.New("review period delay outside the standard range")
	errLockboxDelayMismatch   = errors.New("review period delays differ between chains sharing a lockbox")
)

// LockboxChainDelays holds the review period delays read from one chain that has an ETHLockbox.
type LockboxChainDelays struct {
	ChainID    uint64
	Superchain config.Superchain
	Lockbox    config.ChecksummedAddress
	Delays     report.L1ReviewPeriodReport
}

// ReviewPeriodChecker validates the review period delays of every chain that has an ETHLockbox.
// Each delay must be within the standard range of its superchain, and chains sharing a lockbox
// must use the same delays: a lockbox is only as safe as the shortest delay of any portal that
// can withdraw from it.
type ReviewPeriodChecker struct {
	lgr           log.Logger
	cfgs          []DiskChainConfig
	addrs         config.AddressesJSON
	l1RpcUrls     []string
	superchainIds map[config.Superchain]uint64
}

func NewReviewPeriodChecker(
	lgr log.Logger,
	cfgs []DiskChainConfig,
	addrs config.AddressesJSON,
	l1RpcUrls []string,
	superchainIds map[config.Superchain]uint64,
) *ReviewPeriodChecker {
	return &ReviewPeriodChecker{
		lgr:           lgr,
		cfgs:          cfgs,
		addrs:         addrs,
		l1RpcUrls:     l1RpcUrls,
		superchainIds: superchainIds,
	}
}

func (c *ReviewPeriodChecker) Check(ctx context.Context) error {
	chains, err := c.readDelays(ctx)
	if err != nil {
		return err
	}
	if err := CheckReviewPeriodDelays(chains); err != nil {
		return err
	}
	c.lgr.Info("validated review period delays", "numChains", len(chains))
	return nil
}

// readDelays reads the review period delays of every lockbox chain over the L1 RPC of its superchain.
func (c *ReviewPeriodChecker) readDelays(ctx context.Context) ([]LockboxChainDelays, error) {
	bySuperchain, err := lockboxChains(c.cfgs, c.addrs)
	if err != nil {
		return nil, err
	}

	var out []LockboxChainDelays
	for superchain, chains := range bySuperchain {
		if _, ok := standardConfigParams(superchain); !ok {
			c.lgr.Info("skipping superchain without standard config params", "superchain", superchain)
			continue
		}
		superchainId, ok := c.superchainIds[superchain]
		if !ok {
			return nil, fmt.Errorf("missing superchain chainId for superchain %s", superchain)
		}
		l1RpcUrl, err := config.FindValidL1URL(ctx, c.lgr, c.l1RpcUrls, superchainId)
		if err != nil {
			return nil, fmt.Errorf("missing L1 RPC URL for superchain %s: %w", superchain, err)
		}
		client, err := rpc.DialContext(ctx, l1RpcUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to dial L1 RPC for superchain %s: %w", superchain, err)
		}

		for _, chain := range chains {
			// addresses.json only lists a DelayedWETH for chains with a bisection game, so read it from the
			// SystemConfig instead. Every lockbox chain runs op-contracts/v6.0.0 or later, where the getter exists.
			wethAddr, err := report.DelayedWETHAddress(ctx, client, common.Address(*chain.addrs.SystemConfigProxy))
			if err != nil {
				client.Close()
				return nil, fmt.Errorf("failed to read DelayedWETH address for chain %d: %w", chain.chainID, err)
			}
			delays, err := report.ScanReviewPeriod(ctx, client, common.Address(*chain.addrs.OptimismPortalProxy), wethAddr)
			if err != nil {
				client.Close()
				return nil, fmt.Errorf("failed to read review period delays for chain %d: %w", chain.chainID, err)
			}
			c.lgr.Info("read review period delays",
				"chainID", chain.chainID,
				"lockbox", chain.addrs.EthLockboxProxy,
				"proofMaturityDelaySeconds", delays.ProofMaturityDelaySeconds,
				"disputeGameFinalityDelaySeconds", delays.DisputeGameFinalityDelaySeconds,
				"delayedWETHDelaySeconds", delays.DelayedWETHDelaySeconds,
			)
			out = append(out, LockboxChainDelays{
				ChainID:    chain.chainID,
				Superchain: superchain,
				Lockbox:    *chain.addrs.EthLockboxProxy,
				Delays:     delays,
			})
		}
		client.Close()
	}
	return out, nil
}

type lockboxChain struct {
	chainID uint64
	addrs   *config.AddressesWithRoles
}

// lockboxChains returns, per superchain, the chains whose addresses include an ETHLockbox. Such
// chains must also expose the portal and system config proxies the delays are read from.
func lockboxChains(cfgs []DiskChainConfig, addrs config.AddressesJSON) (map[config.Superchain][]lockboxChain, error) {
	out := make(map[config.Superchain][]lockboxChain)
	for _, cfg := range cfgs {
		chainID := cfg.Config.ChainID
		chainAddrs, ok := addrs[eth.ChainIDFromUInt64(chainID).String()]
		if !ok || chainAddrs == nil || isZeroAddress(chainAddrs.EthLockboxProxy) {
			continue
		}
		for name, addr := range map[string]*config.ChecksummedAddress{
			"OptimismPortalProxy": chainAddrs.OptimismPortalProxy,
			"SystemConfigProxy":   chainAddrs.SystemConfigProxy,
		} {
			if isZeroAddress(addr) {
				return nil, fmt.Errorf("%w: no %s found for lockbox chain %d", errMissingAddress, name, chainID)
			}
		}
		out[cfg.Superchain] = append(out[cfg.Superchain], lockboxChain{chainID: chainID, addrs: chainAddrs})
	}
	return out, nil
}

// CheckReviewPeriodDelays asserts that every delay is within the standard range of its superchain
// and that chains sharing a lockbox have identical delays.
func CheckReviewPeriodDelays(chains []LockboxChainDelays) error {
	byLockbox := make(map[config.ChecksummedAddress]LockboxChainDelays)
	for _, chain := range chains {
		std, ok := standardConfigParams(chain.Superchain)
		if !ok {
			return fmt.Errorf("no standard config params for superchain %s (chain %d)", chain.Superchain, chain.ChainID)
		}

		for _, d := range []struct {
			name  string
			std   validation.Range
			value uint64
		}{
			{"proofMaturityDelaySeconds", std.OptimismPortal2.ProofMaturityDelaySeconds, chain.Delays.ProofMaturityDelaySeconds},
			{"disputeGameFinalityDelaySeconds", std.OptimismPortal2.DisputeGameFinalityDelaySeconds, chain.Delays.DisputeGameFinalityDelaySeconds},
			{"delayedWETHDelaySeconds", std.DelayedWETH.DelaySeconds, chain.Delays.DelayedWETHDelaySeconds},
		} {
			if d.value > math.MaxInt || !d.std.WithinRange(int(d.value)) {
				return fmt.Errorf("%w: chain %d %s is %d, expected within %v",
					errReviewPeriodOutOfRange, chain.ChainID, d.name, d.value, d.std)
			}
		}

		first, seen := byLockbox[chain.Lockbox]
		if !seen {
			byLockbox[chain.Lockbox] = chain
			continue
		}
		if first.Delays != chain.Delays {
			return fmt.Errorf("%w: lockbox %s: chain %d has %+v, chain %d has %+v",
				errLockboxDelayMismatch, chain.Lockbox, first.ChainID, first.Delays, chain.ChainID, chain.Delays)
		}
	}
	return nil
}

func standardConfigParams(superchain config.Superchain) (validation.ConfigParams, bool) {
	switch superchain {
	case config.MainnetSuperchain:
		return validation.StandardConfigParamsMainnet, true
	case config.SepoliaSuperchain:
		return validation.StandardConfigParamsSepolia, true
	default:
		return validation.ConfigParams{}, false
	}
}

func isZeroAddress(addr *config.ChecksummedAddress) bool {
	return addr == nil || *addr == config.ChecksummedAddress{}
}
