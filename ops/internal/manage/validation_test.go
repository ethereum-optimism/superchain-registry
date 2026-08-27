package manage

import (
	"errors"
	"strings"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/stretchr/testify/require"
)

func TestValidateNewChainDataAvailability(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Chain
		wantErr bool
	}{
		{
			name: "ethereum da",
			cfg: config.Chain{
				DataAvailabilityType: "eth-da",
			},
		},
		{
			name: "alternative da type",
			cfg: config.Chain{
				DataAvailabilityType: "alt-da",
			},
			wantErr: true,
		},
		{
			name: "legacy metadata",
			cfg: config.Chain{
				DataAvailabilityType: "eth-da",
				AltDA:                &config.AltDA{},
			},
			wantErr: true,
		},
		{
			name: "legacy challenge address",
			cfg: config.Chain{
				DataAvailabilityType: "eth-da",
				Addresses: config.Addresses{
					DAChallengeAddress: new(config.ChecksummedAddress),
				},
			},
			wantErr: true,
		},
		{
			name:    "missing type",
			cfg:     config.Chain{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNewChainDataAvailability(&tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				require.True(t, errors.Is(err, ErrUnsupportedDataAvailability))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func legacyRegistryFixture(t *testing.T) []DiskChainConfig {
	t.Helper()
	chains := make([]DiskChainConfig, 0, len(legacyAltDAChains))
	for key, chainID := range legacyAltDAChains {
		parts := strings.Split(key, "/")
		require.Len(t, parts, 2)
		chains = append(chains, DiskChainConfig{
			Superchain: config.Superchain(parts[0]),
			ShortName:  parts[1],
			Config: &config.Chain{
				ChainID:              chainID,
				DataAvailabilityType: "alt-da",
			},
		})
	}
	return chains
}

func TestValidateRegistryDataAvailability(t *testing.T) {
	t.Run("existing legacy chains", func(t *testing.T) {
		require.NoError(t, ValidateRegistryDataAvailability(legacyRegistryFixture(t)))
	})

	t.Run("legacy chain cannot migrate without allowlist update", func(t *testing.T) {
		chains := legacyRegistryFixture(t)
		chains[0].Config.DataAvailabilityType = "eth-da"
		require.Error(t, ValidateRegistryDataAvailability(chains))
	})

	t.Run("duplicate legacy chain", func(t *testing.T) {
		chains := legacyRegistryFixture(t)
		chains = append(chains, chains[0])
		require.Error(t, ValidateRegistryDataAvailability(chains))
	})

	tests := []struct {
		name  string
		chain DiskChainConfig
	}{
		{
			name: "new alt-da chain",
			chain: DiskChainConfig{
				Superchain: "mainnet",
				ShortName:  "new-alt-da",
				Config: &config.Chain{
					ChainID:              123456789,
					DataAvailabilityType: "alt-da",
				},
			},
		},
		{
			name: "legacy chain ID under another name",
			chain: DiskChainConfig{
				Superchain: "mainnet",
				ShortName:  "renamed-automata",
				Config: &config.Chain{
					ChainID:              65536,
					DataAvailabilityType: "alt-da",
				},
			},
		},
		{
			name: "new ethereum da chain with legacy metadata",
			chain: DiskChainConfig{
				Superchain: "mainnet",
				ShortName:  "new-eth-da",
				Config: &config.Chain{
					ChainID:              123456790,
					DataAvailabilityType: "eth-da",
					AltDA:                &config.AltDA{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chains := append(legacyRegistryFixture(t), tt.chain)
			require.Error(t, ValidateRegistryDataAvailability(chains))
		})
	}
}

func TestValidateGenesisIntegrity(t *testing.T) {
	t.Run("validates genesis successfully", func(t *testing.T) {
		t.Parallel()

		cfg, err := ReadChainConfig("testdata", "sepolia", "testchain")
		require.NoError(t, err)

		genesis, err := ReadSuperchainGenesis("testdata", "sepolia", "testchain")
		require.NoError(t, err)

		err = ValidateGenesisIntegrity(cfg, genesis)
		require.NoError(t, err)
	})

	type testCase struct {
		name    string
		mutator func(*core.Genesis, *config.Chain)
	}

	tests := []testCase{
		{
			name: "fails when the hash is wrong in the config",
			mutator: func(_ *core.Genesis, c *config.Chain) {
				c.Genesis.L2.Hash = common.HexToHash("0x1234")
			},
		},
		{
			name: "fails when timestamp is modified",
			mutator: func(g *core.Genesis, _ *config.Chain) {
				g.Timestamp++
			},
		},
		{
			name: "fails when nonce is modified",
			mutator: func(g *core.Genesis, _ *config.Chain) {
				g.Nonce++
			},
		},
		{
			name: "fails when extradata is modified",
			mutator: func(g *core.Genesis, _ *config.Chain) {
				g.ExtraData = append(g.ExtraData, byte(0x00))
			},
		},
		{
			name: "fails when gas limit is modified",
			mutator: func(g *core.Genesis, _ *config.Chain) {
				g.GasLimit++
			},
		},
		{
			name: "fails when base fee is modified",
			mutator: func(g *core.Genesis, _ *config.Chain) {
				g.BaseFee.Add(g.BaseFee, common.Big1)
			},
		},
		{
			name: "fails when blob gas parameters are modified",
			mutator: func(g *core.Genesis, _ *config.Chain) {
				g.ExcessBlobGas = new(uint64)
				*g.ExcessBlobGas = 1
			},
		},
		{
			name: "fails when state hash is modified",
			mutator: func(g *core.Genesis, _ *config.Chain) {
				g.StateHash = new(common.Hash)
				g.StateHash[0] = 0x01
				g.Alloc = nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := ReadChainConfig("testdata", "sepolia", "testchain")
			require.NoError(t, err)

			genesis, err := ReadSuperchainGenesis("testdata", "sepolia", "testchain")
			require.NoError(t, err)

			tt.mutator(genesis, cfg)

			err = ValidateGenesisIntegrity(cfg, genesis)
			require.Error(t, err)
			require.ErrorContains(t, err, "genesis hash mismatch")
		})
	}
}
