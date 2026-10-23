package manage

import (
	"context"
	"log/slog"
	"testing"

	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/testutil/mockrpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

func TestScanReviewPeriodDelays(t *testing.T) {
	mockRPC := mockrpc.NewMockRPC(
		t,
		testlog.Logger(t, slog.LevelWarn),
		mockrpc.WithExpectationsFile(t, "testdata/review_period_scan.json"),
	)
	client, err := rpc.Dial(mockRPC.Endpoint())
	require.NoError(t, err)

	delays, err := scanReviewPeriodDelays(
		context.Background(),
		client,
		common.HexToAddress("0x034edD2A225f7f429A63E0f1D2084B9E0A93b538"),
		common.HexToAddress("0x93dDE1822EfF3c3Bcdd7446c3E815EcF6952944c"),
	)
	require.NoError(t, err)
	require.Equal(t, ReviewPeriodDelays{
		ProofMaturityDelaySeconds:       604800,
		DisputeGameFinalityDelaySeconds: 302400,
		DelayedWETHDelaySeconds:         302400,
	}, delays)
	mockRPC.AssertExpectations(t)
}

func TestCheckReviewPeriodDelays(t *testing.T) {
	lockboxA := config.ChecksummedAddress(common.HexToAddress("0xaaaa"))
	lockboxB := config.ChecksummedAddress(common.HexToAddress("0xbbbb"))
	standard := ReviewPeriodDelays{
		ProofMaturityDelaySeconds:       604800,
		DisputeGameFinalityDelaySeconds: 302400,
		DelayedWETHDelaySeconds:         302400,
	}
	shortened := ReviewPeriodDelays{
		ProofMaturityDelaySeconds:       86400,
		DisputeGameFinalityDelaySeconds: 43200,
		DelayedWETHDelaySeconds:         43200,
	}
	chain := func(id uint64, lockbox config.ChecksummedAddress, delays ReviewPeriodDelays) LockboxChainDelays {
		return LockboxChainDelays{ChainID: id, Superchain: config.MainnetSuperchain, Lockbox: lockbox, Delays: delays}
	}

	t.Run("in range and consistent per lockbox", func(t *testing.T) {
		require.NoError(t, CheckReviewPeriodDelays([]LockboxChainDelays{
			chain(1, lockboxA, standard),
			chain(2, lockboxA, standard),
			chain(3, lockboxB, shortened),
		}))
	})

	t.Run("out of range", func(t *testing.T) {
		tooShort := standard
		tooShort.ProofMaturityDelaySeconds = 3600
		err := CheckReviewPeriodDelays([]LockboxChainDelays{chain(1, lockboxA, tooShort)})
		require.ErrorIs(t, err, errReviewPeriodOutOfRange)
		require.ErrorContains(t, err, "chain 1 proofMaturityDelaySeconds is 3600")

		tooLong := standard
		tooLong.DelayedWETHDelaySeconds = 604801
		err = CheckReviewPeriodDelays([]LockboxChainDelays{chain(1, lockboxA, tooLong)})
		require.ErrorIs(t, err, errReviewPeriodOutOfRange)
		require.ErrorContains(t, err, "delayedWETHDelaySeconds is 604801")
	})

	t.Run("same lockbox address on another superchain is a different lockbox", func(t *testing.T) {
		sepolia := chain(2, lockboxA, shortened)
		sepolia.Superchain = config.SepoliaSuperchain
		require.NoError(t, CheckReviewPeriodDelays([]LockboxChainDelays{chain(1, lockboxA, standard), sepolia}))
	})

	t.Run("mismatch within a lockbox", func(t *testing.T) {
		err := CheckReviewPeriodDelays([]LockboxChainDelays{
			chain(1, lockboxA, standard),
			chain(2, lockboxA, shortened),
		})
		require.ErrorIs(t, err, errLockboxDelayMismatch)
		require.ErrorContains(t, err, "chain 1")
		require.ErrorContains(t, err, "chain 2")
	})

	t.Run("unsupported superchain", func(t *testing.T) {
		c := chain(1, lockboxA, standard)
		c.Superchain = "sepolia-devnet-2"
		require.Error(t, CheckReviewPeriodDelays([]LockboxChainDelays{c}))
	})
}

func TestLockboxChains(t *testing.T) {
	addrs := loadAddresses(t, validAddressesPath)
	cfgs := func(ids ...uint64) []DiskChainConfig {
		out := make([]DiskChainConfig, 0, len(ids))
		for _, id := range ids {
			out = append(out, DiskChainConfig{Superchain: config.MainnetSuperchain, Config: &config.Chain{ChainID: id}})
		}
		return out
	}

	t.Run("chains without a lockbox are skipped", func(t *testing.T) {
		addrs["903"] = &config.AddressesWithRoles{}
		groups, err := lockboxChains(cfgs(903), addrs)
		require.NoError(t, err)
		require.Empty(t, groups)
	})

	t.Run("superchains without standard config params are skipped before address checks", func(t *testing.T) {
		devnet := cfgs(901)
		devnet[0].Superchain = "sepolia-devnet-2"
		groups, err := lockboxChains(devnet, addrs)
		require.NoError(t, err)
		require.Empty(t, groups)
	})

	t.Run("lockbox chains must expose the delay holders", func(t *testing.T) {
		// The depset fixtures only carry lockbox and factory addresses.
		_, err := lockboxChains(cfgs(901), addrs)
		require.ErrorIs(t, err, errMissingAddress)
	})

	t.Run("lockbox chains are grouped by superchain", func(t *testing.T) {
		full := config.ChecksummedAddress(common.HexToAddress("0x1234"))
		addrs["901"].OptimismPortalProxy = &full
		addrs["901"].SystemConfigProxy = &full
		groups, err := lockboxChains(cfgs(901, 903), addrs)
		require.NoError(t, err)
		require.Len(t, groups[config.MainnetSuperchain], 1)
		require.EqualValues(t, 901, groups[config.MainnetSuperchain][0].chainID)
	})
}
