package manage

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/report"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3"
)

var (
	errReviewPeriodOutOfRange = errors.New("review period delay outside the standard range")
	errLockboxDelayMismatch   = errors.New("review period delays differ between chains sharing a lockbox")

	delayedWETHABI                     = w3.MustNewFunc("delayedWETH()", "address")
	proofMaturityDelaySecondsABI       = w3.MustNewFunc("proofMaturityDelaySeconds()", "uint256")
	disputeGameFinalityDelaySecondsABI = w3.MustNewFunc("disputeGameFinalityDelaySeconds()", "uint256")
	delayABI                           = w3.MustNewFunc("delay()", "uint256")
)

// ReviewPeriodDelays holds the delays that gate withdrawals: the proof maturity delay on the
// OptimismPortal, the dispute game finality delay on the AnchorStateRegistry and the DelayedWETH
// withdrawal delay.
type ReviewPeriodDelays struct {
	ProofMaturityDelaySeconds       uint64
	DisputeGameFinalityDelaySeconds uint64
	DelayedWETHDelaySeconds         uint64
}

// LockboxChainDelays holds the review period delays read from one chain that has an ETHLockbox,
// together with the standard config params of its superchain.
type LockboxChainDelays struct {
	ChainID    uint64
	Superchain config.Superchain
	Lockbox    config.ChecksummedAddress
	Standard   validation.ConfigParams
	Delays     ReviewPeriodDelays
}

// ReviewPeriodChecker validates the review period delays of every chain that has an ETHLockbox.
// Each delay must be within the standard range of its superchain, and chains sharing a lockbox
// must use the same delays. The AnchorStateRegistry and DelayedWETH are shared by chains in an
// interop set, so their delays are equal by construction. The proof maturity delay is per portal and
// the contracts allow it to differ. Keeping it uniform across chains in an interop set is a policy
// enforced off-chain, so a single chain cannot shorten the review period of the entire set.
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
	for superchain, group := range bySuperchain {
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

		for _, chain := range group.chains {
			delays, err := scanReviewPeriodDelays(
				ctx,
				client,
				common.Address(*chain.addrs.OptimismPortalProxy),
				common.Address(*chain.addrs.SystemConfigProxy),
			)
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
				Standard:   group.standard,
				Delays:     delays,
			})
		}
		client.Close()
	}
	return out, nil
}

// scanReviewPeriodDelays reads a chain's review period delays from its OptimismPortal and SystemConfig
// proxies. The DelayedWETH is resolved through `SystemConfig.delayedWETH()`
func scanReviewPeriodDelays(
	ctx context.Context,
	client *rpc.Client,
	portalAddr common.Address,
	systemConfigAddr common.Address,
) (ReviewPeriodDelays, error) {
	w3Client := w3.NewClient(client)
	var delays ReviewPeriodDelays

	var wethAddr common.Address
	if err := report.CallBatch(ctx, w3Client, batchCall(systemConfigAddr, delayedWETHABI, &wethAddr)); err != nil {
		return delays, fmt.Errorf("failed to get DelayedWETH address: %w", err)
	}

	var proofMaturity, finalityDelay, wethDelay big.Int
	if err := report.CallBatch(
		ctx,
		w3Client,
		batchCall(portalAddr, proofMaturityDelaySecondsABI, &proofMaturity),
		batchCall(portalAddr, disputeGameFinalityDelaySecondsABI, &finalityDelay),
		batchCall(wethAddr, delayABI, &wethDelay),
	); err != nil {
		return delays, fmt.Errorf("failed to get review period delays: %w", err)
	}

	for _, v := range []struct {
		name string
		val  *big.Int
		out  *uint64
	}{
		{"proof maturity delay", &proofMaturity, &delays.ProofMaturityDelaySeconds},
		{"dispute game finality delay", &finalityDelay, &delays.DisputeGameFinalityDelaySeconds},
		{"DelayedWETH delay", &wethDelay, &delays.DelayedWETHDelaySeconds},
	} {
		if !v.val.IsUint64() {
			return delays, fmt.Errorf("unexpectedly large %s: %s", v.name, v.val.String())
		}
		*v.out = v.val.Uint64()
	}

	return delays, nil
}

func batchCall(to common.Address, fn *w3.Func, out any) report.BatchCall {
	return report.BatchCall{
		To:      to,
		Encoder: func() ([]byte, error) { return fn.EncodeArgs() },
		Decoder: func(rawOutput []byte) error { return fn.DecodeReturns(rawOutput, out) },
	}
}

type lockboxChain struct {
	chainID uint64
	addrs   *config.AddressesWithRoles
}

// lockboxGroup is the set of lockbox chains of one superchain and the standard params they are
// checked against.
type lockboxGroup struct {
	standard validation.ConfigParams
	chains   []lockboxChain
}

// lockboxChains returns, per superchain, the chains whose addresses include an ETHLockbox.
// Lockbox chains must also expose the portal and system config proxies the delays are read from.
func lockboxChains(cfgs []DiskChainConfig, addrs config.AddressesJSON) (map[config.Superchain]*lockboxGroup, error) {
	out := make(map[config.Superchain]*lockboxGroup)
	for _, cfg := range cfgs {
		standard, ok := standardConfigParams(cfg.Superchain)
		if !ok {
			continue
		}
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
		group, ok := out[cfg.Superchain]
		if !ok {
			group = &lockboxGroup{standard: standard}
			out[cfg.Superchain] = group
		}
		group.chains = append(group.chains, lockboxChain{chainID: chainID, addrs: chainAddrs})
	}
	return out, nil
}

// CheckReviewPeriodDelays asserts that every delay is within the standard range carried by its
// entry and that chains sharing a lockbox have identical delays, including the per-portal proof
// maturity delay.
func CheckReviewPeriodDelays(chains []LockboxChainDelays) error {
	// Proxies are deployed with CREATE2, so the same lockbox address can exist on several L1s.
	type lockboxKey struct {
		superchain config.Superchain
		lockbox    config.ChecksummedAddress
	}
	byLockbox := make(map[lockboxKey]LockboxChainDelays)
	for _, chain := range chains {
		std := chain.Standard
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

		key := lockboxKey{superchain: chain.Superchain, lockbox: chain.Lockbox}
		first, seen := byLockbox[key]
		if !seen {
			byLockbox[key] = chain
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
