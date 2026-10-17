package manage

import (
	"testing"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestDiffChainAgainstOnchain(t *testing.T) {
	addr := func(b byte) *config.ChecksummedAddress {
		return config.NewChecksummedAddress(common.Address{b})
	}

	onchain := config.AddressesWithRoles{
		Addresses: config.Addresses{
			SystemConfigProxy:       addr(0x01),
			OptimismPortalProxy:     addr(0x02),
			DisputeGameFactoryProxy: config.NewChecksummedAddress(common.Address{}),
		},
		Roles: config.Roles{
			ProxyAdminOwner: addr(0x10),
			Guardian:        addr(0x11),
		},
	}

	t.Run("matching fields", func(t *testing.T) {
		disk := &config.Chain{
			Addresses: config.Addresses{SystemConfigProxy: addr(0x01)},
			Roles:     config.Roles{ProxyAdminOwner: addr(0x10)},
		}
		require.Empty(t, DiffChainAgainstOnchain(disk, onchain))
	})

	t.Run("fields absent from toml are skipped", func(t *testing.T) {
		require.Empty(t, DiffChainAgainstOnchain(&config.Chain{}, onchain))
	})

	t.Run("mismatched fields", func(t *testing.T) {
		disk := &config.Chain{
			Addresses: config.Addresses{
				SystemConfigProxy:       addr(0x01),
				OptimismPortalProxy:     addr(0x99),
				DisputeGameFactoryProxy: addr(0x03),
			},
			Roles: config.Roles{
				ProxyAdminOwner: addr(0x98),
				Guardian:        addr(0x11),
			},
		}
		mismatches := DiffChainAgainstOnchain(disk, onchain)
		require.Equal(t, []OnchainMismatch{
			{Section: "roles", Field: "ProxyAdminOwner", Disk: *addr(0x98), Onchain: *addr(0x10)},
			{Section: "addresses", Field: "DisputeGameFactoryProxy", Disk: *addr(0x03), Onchain: config.ChecksummedAddress{}},
			{Section: "addresses", Field: "OptimismPortalProxy", Disk: *addr(0x99), Onchain: *addr(0x02)},
		}, mismatches)
		require.False(t, mismatches[0].Unresolved())
		require.True(t, mismatches[1].Unresolved())
		require.False(t, mismatches[2].Unresolved())
		require.Equal(t,
			"[roles] ProxyAdminOwner: toml=0x9800000000000000000000000000000000000000 onchain=0x1000000000000000000000000000000000000000",
			mismatches[0].String(),
		)
	})
}
