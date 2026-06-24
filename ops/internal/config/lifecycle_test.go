package config

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// testNow is the reference time for lifecycle tests. The hardfork activations in
// baseChain (1000, 2000) are in the past relative to it; activations above it are in
// the future.
const testNow uint64 = 10000

// baseChain returns a minimal but representative chain config for lifecycle tests.
func baseChain() *Chain {
	addr := NewChecksummedAddress(common.HexToAddress("0x1111111111111111111111111111111111111111"))
	return &Chain{
		Name:              "Test Chain",
		PublicRPC:         "https://rpc.example",
		ChainID:           42,
		BlockTime:         2,
		SeqWindowSize:     3600,
		MaxSequencerDrift: 600,
		Hardforks: Hardforks{
			CanyonTime: NewHardforkTime(1000),
			DeltaTime:  NewHardforkTime(2000),
		},
		Genesis: Genesis{
			L2Time: 500,
			L2:     GenesisRef{Hash: common.HexToHash("0xabc"), Number: 1},
			SystemConfig: SystemConfig{
				BatcherAddr: *addr,
				GasLimit:    30000000,
			},
		},
		Roles:     Roles{ProxyAdminOwner: addr},
		Addresses: Addresses{SystemConfigProxy: addr},
	}
}

func TestCheckImmutableFields(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(c *Chain)
		wantErr bool
	}{
		{
			name:   "no change is allowed",
			mutate: func(c *Chain) {},
		},
		{
			name: "mutable fields may change freely",
			mutate: func(c *Chain) {
				c.Name = "Renamed"
				c.PublicRPC = "https://new.example"
				c.Roles.ProxyAdminOwner = NewChecksummedAddress(common.HexToAddress("0x2222222222222222222222222222222222222222"))
				c.Addresses.SystemConfigProxy = nil
			},
		},
		{
			name:    "chain_id is immutable",
			mutate:  func(c *Chain) { c.ChainID = 43 },
			wantErr: true,
		},
		{
			name:    "block_time is immutable",
			mutate:  func(c *Chain) { c.BlockTime = 1 },
			wantErr: true,
		},
		{
			name:    "genesis system_config is immutable",
			mutate:  func(c *Chain) { c.Genesis.SystemConfig.GasLimit = 60000000 },
			wantErr: true,
		},
		{
			name:    "genesis l2 ref is immutable",
			mutate:  func(c *Chain) { c.Genesis.L2.Hash = common.HexToHash("0xdef") },
			wantErr: true,
		},
		{
			name:   "appending a new hardfork is allowed",
			mutate: func(c *Chain) { c.Hardforks.EcotoneTime = NewHardforkTime(3000) },
		},
		{
			name:    "changing a past hardfork activation is rejected",
			mutate:  func(c *Chain) { c.Hardforks.CanyonTime = NewHardforkTime(1500) },
			wantErr: true,
		},
		{
			name:    "removing a past hardfork activation is rejected",
			mutate:  func(c *Chain) { c.Hardforks.DeltaTime = nil },
			wantErr: true,
		},
		{
			name: "scheduling a future hardfork activation is allowed",
			mutate: func(c *Chain) {
				c.Hardforks.EcotoneTime = NewHardforkTime(testNow + 1000)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			old := baseChain()
			next := baseChain()
			tt.mutate(next)

			err := CheckImmutableFields(old, next, testNow)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestRescheduleFutureHardfork covers adjusting an already-set future activation,
// which requires a non-default old version, so it lives outside the table.
func TestRescheduleFutureHardfork(t *testing.T) {
	t.Parallel()

	old := baseChain()
	old.Hardforks.EcotoneTime = NewHardforkTime(testNow + 1000) // future

	next := baseChain()
	next.Hardforks.EcotoneTime = NewHardforkTime(testNow + 5000) // pushed out

	require.NoError(t, CheckImmutableFields(old, next, testNow),
		"pushing out a not-yet-active hardfork must be allowed")
}

// TestFreezeJustActivatedHardfork ensures an activation at exactly now is frozen.
func TestFreezeJustActivatedHardfork(t *testing.T) {
	t.Parallel()

	old := baseChain()
	old.Hardforks.EcotoneTime = NewHardforkTime(testNow) // activates at now

	next := baseChain()
	next.Hardforks.EcotoneTime = NewHardforkTime(testNow + 1) // attempt to move it

	require.Error(t, CheckImmutableFields(old, next, testNow),
		"an activation at or before now must be frozen")
}

// TestEveryChainFieldHasLifecycle guards against silently adding a field without
// classifying it, which would leave it unenforced.
func TestEveryChainFieldHasLifecycle(t *testing.T) {
	old := baseChain()
	next := baseChain()
	// Sanity: the comparator runs without panicking over the full struct.
	require.NoError(t, CheckImmutableFields(old, next, testNow))

	tp := reflect.TypeOf(*old)
	for i := 0; i < tp.NumField(); i++ {
		field := tp.Field(i)
		lc := FieldLifecycle(field.Tag.Get(lifecycleTag))
		require.Containsf(t,
			[]FieldLifecycle{LifecycleImmutable, LifecycleAppendOnly, LifecycleMutable},
			lc,
			"Chain field %q is missing a valid `lifecycle` tag (got %q)", field.Name, lc,
		)
	}
}
