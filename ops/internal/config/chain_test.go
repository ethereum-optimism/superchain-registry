package config

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/optimism/op-chain-ops/interopgen"
	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch/script"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestSuperchainLevel_Marshaling(t *testing.T) {
	type holder struct {
		Lvl SuperchainLevel `toml:"lvl"`
	}
	tests := []struct {
		level SuperchainLevel
		exp   string
	}{
		{SuperchainLevelNonStandard, "lvl = 0"},
		{SuperchainLevelStandardCandidate, "lvl = 1"},
		{SuperchainLevelStandard, "lvl = 2"},
	}
	for _, tt := range tests {
		data, err := toml.Marshal(holder{tt.level})
		require.NoError(t, err)
		require.Equal(t, tt.exp, strings.TrimSpace(string(data)))

		var h holder
		require.NoError(t, toml.Unmarshal(data, &h))
		require.Equal(t, tt.level, h.Lvl)
	}

	errTests := []struct {
		name string
		in   string
	}{
		{"invalid float type", "lvl = 1.0"},
		{"invalid string type", "lvl = \"sup\""},
		{"invalid hex type", "lvl = 0x123"},
	}
	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			var h holder
			err := toml.Unmarshal([]byte(tt.in), &h)
			require.Error(t, err)
		})
	}
}

func TestChain_Marshaling(t *testing.T) {
	data, err := os.ReadFile("testdata/all.toml")
	require.NoError(t, err)
	var chain Chain
	require.NoError(t, toml.Unmarshal(data, &chain))

	var buf bytes.Buffer
	require.NoError(t, toml.NewEncoder(&buf).Encode(chain))
	require.Equal(t, string(data), buf.String())
}

func TestAddressesWithRoles_Marshaling(t *testing.T) {
	testData := AddressesWithRoles{
		Addresses: script.Addresses{
			L2OpchainDeployment: interopgen.L2OpchainDeployment{
				AddressManager:                     common.Address{'A'},
				L1CrossDomainMessengerProxy:        common.Address{'B'},
				L1ERC721BridgeProxy:                common.Address{'C'},
				OptimismMintableERC20FactoryProxy:  common.Address{'D'},
				SystemConfigProxy:                  common.Address{'E'},
				L1StandardBridgeProxy:              common.Address{'F'},
				OptimismPortalProxy:                common.Address{'G'},
				OpChainProxyAdmin:                  common.Address{'H'},
				AnchorStateRegistryProxy:           common.Address{'I'},
				DelayedWETHPermissionedGameProxy:   common.Address{'J'},
				DelayedWETHPermissionlessGameProxy: common.Address{'K'},
				DisputeGameFactoryProxy:            common.Address{'L'},
				FaultDisputeGame:                   common.Address{'M'},
				PermissionedDisputeGame:            common.Address{'O'},
			},
			Mips:                common.Address{'P'},
			SuperchainConfig:    common.Address{'Q'},
			PreimageOracle:      common.Address{'R'},
			L2OutputOracleProxy: common.Address{'S'},
		},
		Roles: script.Roles{
			SystemConfigOwner:      common.Address{'T'},
			OpChainProxyAdminOwner: common.Address{'U'},
			Guardian:               common.Address{'V'},
			Challenger:             common.Address{'W'},
			UnsafeBlockSigner:      common.Address{'X'},
			BatchSubmitter:         common.Address{'Y'},
		},
	}

	expData, err := os.ReadFile("testdata/expected-addresses-with-roles.json")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, json.NewEncoder(&buf).Encode(testData))
	require.JSONEq(t, string(expData), buf.String())
}
