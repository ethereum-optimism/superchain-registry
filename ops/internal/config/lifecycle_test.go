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
		BatchInboxAddr:    addr,
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
			name: "batch_inbox_addr is immutable",
			mutate: func(c *Chain) {
				c.BatchInboxAddr = NewChecksummedAddress(common.HexToAddress("0x3333333333333333333333333333333333333333"))
			},
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
		{
			name: "adding the optional pectra_blob_schedule_time is allowed",
			mutate: func(c *Chain) {
				c.Hardforks.PectraBlobScheduleTime = NewHardforkTime(1500)
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

// TestFreezePastBlobSchedule ensures the optional pectra_blob_schedule_time, once its
// activation is in the past, is frozen like any other hardfork activation.
func TestFreezePastBlobSchedule(t *testing.T) {
	t.Parallel()

	old := baseChain()
	old.Hardforks.PectraBlobScheduleTime = NewHardforkTime(1500) // in the past

	next := baseChain()
	next.Hardforks.PectraBlobScheduleTime = NewHardforkTime(1600) // attempt to move it

	require.Error(t, CheckImmutableFields(old, next, testNow),
		"a blob-schedule activation already in the past must be frozen")
}

// TestKeepKarstUpgradeGas covers the non-timestamp hardfork modifier. It is governed
// by karst_time: freely changeable until Karst is in the past, frozen afterwards. A
// bool has no "unset" state, so the freeze is driven by Karst's activation, not by the
// flag's own value (false->true after activation must be rejected just like true->false).
func TestKeepKarstUpgradeGas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		karstTime *HardforkTime
		oldFlag   bool
		newFlag   bool
		wantErr   bool
	}{
		{name: "karst unset, flag may be set", karstTime: nil, oldFlag: false, newFlag: true},
		{name: "karst in future, flag may be set", karstTime: NewHardforkTime(testNow + 1000), oldFlag: false, newFlag: true},
		{name: "karst in future, flag may be cleared", karstTime: NewHardforkTime(testNow + 1000), oldFlag: true, newFlag: false},
		{name: "karst in past, flag frozen (cleared)", karstTime: NewHardforkTime(1000), oldFlag: true, newFlag: false, wantErr: true},
		{name: "karst in past, flag frozen (set)", karstTime: NewHardforkTime(1000), oldFlag: false, newFlag: true, wantErr: true},
		{name: "karst in past, flag unchanged", karstTime: NewHardforkTime(1000), oldFlag: true, newFlag: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			old := baseChain()
			old.Hardforks.KarstTime = tt.karstTime
			old.Hardforks.KeepKarstUpgradeGas = tt.oldFlag

			next := baseChain()
			next.Hardforks.KarstTime = tt.karstTime
			next.Hardforks.KeepKarstUpgradeGas = tt.newFlag

			err := CheckImmutableFields(old, next, testNow)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestHardforksFieldsClassified guards against silently adding a field to Hardforks
// without teaching checkAppendOnly how it ages: every field must be either a
// *HardforkTime activation or a registered non-timestamp modifier.
func TestHardforksFieldsClassified(t *testing.T) {
	tp := reflect.TypeOf(Hardforks{})
	for i := 0; i < tp.NumField(); i++ {
		field := tp.Field(i)
		classified := isHardforkTime(field.Type) || hardforkModifiers[field.Name] != ""
		require.Truef(t, classified,
			"Hardforks field %q (%s) is unclassified: add a *HardforkTime activation or register it in hardforkModifiers",
			field.Name, field.Type)
	}
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
