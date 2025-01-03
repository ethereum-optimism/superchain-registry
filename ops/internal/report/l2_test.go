package report

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"math/big"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-chain-ops/foundry"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/artifacts"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

type artifactsProvider struct {
	fs      foundry.StatDirFs
	mtx     sync.Mutex
	cleanup func() error
	t       *testing.T
}

func newArtifactsProvider(t *testing.T) *artifactsProvider {
	prov := &artifactsProvider{
		t: t,
		cleanup: func() error {
			return nil
		},
	}
	t.Cleanup(func() {
		require.NoError(t, prov.cleanup())
	})
	return prov
}

func (a *artifactsProvider) Artifacts(ctx context.Context, loc *artifacts.Locator) (foundry.StatDirFs, error) {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	if a.fs != nil {
		return a.fs, nil
	}

	afacts, cleanup, err := artifacts.Download(ctx, loc, artifacts.LogProgressor(testlog.Logger(a.t, log.LevelInfo)))
	require.NoError(a.t, err)
	a.cleanup = cleanup
	a.fs = afacts
	return a.fs, nil
}

func TestScanL2(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*state.State) *state.State
		wantErr    string
		wantReport L2Report
	}{
		{
			name: "successful scan",
			setup: func(s *state.State) *state.State {
				return s
			},
			wantReport: L2Report{
				Release:             string(validation.Semver170),
				ProvidedGenesisHash: common.HexToHash("0xcd901673f97d59259fa09b0b01b8787f5d25d9f1808566990673519be65cc3ae"),
				StandardGenesisHash: common.HexToHash("0xcd901673f97d59259fa09b0b01b8787f5d25d9f1808566990673519be65cc3ae"),
				AccountDiffs:        []AccountDiff{},
			},
		},
		{
			name: "nil intent",
			setup: func(s *state.State) *state.State {
				s.AppliedIntent = nil
				return s
			},
			wantErr: "no intent found in original state",
		},
		{
			name: "non-tag L2 contracts locator",
			setup: func(s *state.State) *state.State {
				s.AppliedIntent.L2ContractsLocator = artifacts.MustNewLocatorFromURL("https://example.com")
				return s
			},
			wantErr: "must use a tag for L2 contracts locator",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			originalState := readState(t)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			st := tt.setup(originalState)
			report, err := ScanL2(ctx, st)

			if tt.wantErr == "" {
				require.NoError(t, err)
				require.Equal(t, tt.wantReport, *report)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
				require.Nil(t, report)
			}
		})
	}
}

func TestDiffL2Genesis(t *testing.T) {
	prov := newArtifactsProvider(t)

	testAddr := common.HexToAddress("0x4200000000000000000000000000000000000000")
	tests := []struct {
		name      string
		setup     func(*state.State) *state.State
		wantErr   string
		wantDiffs []AccountDiff
	}{
		{
			name: "successful scan",
			setup: func(s *state.State) *state.State {
				return s
			},
			wantDiffs: []AccountDiff{},
		},
		{
			name: "nil intent",
			setup: func(s *state.State) *state.State {
				s.AppliedIntent = nil
				return s
			},
			wantErr: "no intent found in original state",
		},
		{
			name: "multiple chains in intent",
			setup: func(s *state.State) *state.State {
				s.AppliedIntent.Chains = append(s.AppliedIntent.Chains, s.AppliedIntent.Chains[0])
				return s
			},
			wantErr: "expected exactly one chain in original intent, got 2",
		},
		{
			name: "multiple chains in state",
			setup: func(s *state.State) *state.State {
				s.Chains = append(s.Chains, s.Chains[0])
				return s
			},
			wantErr: "expected exactly one chain in original state, got 2",
		},
		{
			name: "unsupported l1 chain",
			setup: func(s *state.State) *state.State {
				s.AppliedIntent.L1ChainID = 999
				return s
			},
			wantErr: "unsupported L1 chain ID: 999",
		},
		{
			name: "different account balance",
			setup: func(s *state.State) *state.State {
				account := s.Chains[0].Allocs.Data.Accounts[testAddr]
				*account.Balance = *big.NewInt(999)
				return s
			},
			wantDiffs: []AccountDiff{
				{
					Address:        testAddr,
					Added:          false,
					Removed:        false,
					BalanceChanged: true,

					OldBalance: big.NewInt(0),
					NewBalance: big.NewInt(999),
				},
			},
		},
		{
			name: "additional account in original",
			setup: func(s *state.State) *state.State {
				s.Chains[0].Allocs.Data.Accounts[common.HexToAddress("0x111")] = types.Account{
					Balance: big.NewInt(999),
					Code:    []byte{0x01},
					Nonce:   1,
				}
				return s
			},
			wantDiffs: []AccountDiff{
				{
					Address:        common.HexToAddress("0x111"),
					Added:          true,
					CodeChanged:    true,
					BalanceChanged: true,
					NonceChanged:   true,
					NewCode:        []byte{0x01},
					NewBalance:     big.NewInt(999),
					NewNonce:       1,
				},
			},
		},
		{
			name: "different account code",
			setup: func(s *state.State) *state.State {
				account := s.Chains[0].Allocs.Data.Accounts[testAddr]
				s.Chains[0].Allocs.Data.Accounts[testAddr] = types.Account{
					Balance: account.Balance,
					Code:    []byte{0x42, 0x42},
					Nonce:   account.Nonce,
					Storage: account.Storage,
				}
				return s
			},
			wantDiffs: []AccountDiff{
				{
					Address:     testAddr,
					Added:       false,
					Removed:     false,
					CodeChanged: true,
					OldCode:     readTestData(t, "proxy.bin"),
					NewCode:     []byte{0x42, 0x42},
				},
			},
		},
		{
			name: "different storage values",
			setup: func(s *state.State) *state.State {
				addr := common.HexToAddress("0x4200000000000000000000000000000000000043")
				account := s.Chains[0].Allocs.Data.Accounts[addr]
				account.Storage[common.HexToHash("0xabcd")] = common.HexToHash("0x1234")
				return s
			},
			wantDiffs: []AccountDiff{
				{
					Address: common.HexToAddress("0x4200000000000000000000000000000000000043"),

					StorageChanges: []StorageDiff{
						{
							Key:      common.HexToHash("0xabcd"),
							Added:    true,
							Removed:  false,
							NewValue: common.HexToHash("0x1234"),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			originalState := readState(t)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			st := tt.setup(originalState)
			_, diffs, err := DiffL2Genesis(ctx, prov, st)

			if tt.wantErr == "" {
				require.NoError(t, err)
				require.Equal(t, tt.wantDiffs, diffs)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
				require.Nil(t, diffs)
			}
		})
	}
}

func readState(t *testing.T) *state.State {
	f, err := os.OpenFile("testdata/deployer-state.json.gz", os.O_RDONLY, 0)
	require.NoError(t, err)
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	require.NoError(t, err)
	defer gzr.Close()

	var s state.State
	require.NoError(t, json.NewDecoder(gzr).Decode(&s))
	return &s
}

func readTestData(t *testing.T, fname string) []byte {
	data, err := os.ReadFile(path.Join("testdata", fname))
	require.NoError(t, err)
	return data
}
