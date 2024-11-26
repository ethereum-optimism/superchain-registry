package validation

import (
	"testing"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"
)

func TestCheckIsGloballyUnique(t *testing.T) {
	tt := []struct {
		testName    string
		chain       superchain.ChainConfig
		expectedErr error
	}{
		{
			"Success",
			superchain.ChainConfig{
				ChainID:   10,
				Name:      "OP Mainnet",
				PublicRPC: "https://mainnet.optimism.io",
			},
			nil,
		},
		{
			"ErrChainIdNotListed",
			superchain.ChainConfig{
				ChainID:   uint64(12309845720398547027),
				Name:      "testuniquechain",
				PublicRPC: "http://fakeurl.com",
			},
			ErrChainIdNotListed,
		},
		{
			"ErrLocalChainNameMismatch",
			superchain.ChainConfig{
				ChainID:   10,
				Name:      "oops mainnet",
				PublicRPC: "http://fakeurl.com",
			},
			ErrLocalChainNameMismatch,
		},
		{
			"ErrChainPublicRpcNotListed",
			superchain.ChainConfig{
				ChainID:   10,
				Name:      "OP Mainnet",
				PublicRPC: "https://mainnet-fake.optimism.io",
			},
			ErrChainPublicRpcNotListed,
		},
		{
			"ErrChainIdDuplicated",
			superchain.ChainConfig{
				ChainID:   10,
				Name:      "OP Mainnet",
				PublicRPC: "https://mainnet.optimism.io",
			},
			ErrChainIdDuplicated,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			err := checkIsGloballyUnique(globalChainIds, localChains, &tc.chain)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tt := []struct {
		input string
		want  string
	}{
		{"https://rpc.zora.energy/", "https://rpc.zora.energy/"},
		{"https://rpc.zora.energy", "https://rpc.zora.energy/"},
	}

	for _, tc := range tt {
		got, err := normalizeURL(tc.input)
		require.NoError(t, err)
		require.Equal(t, tc.want, got)
	}
}
