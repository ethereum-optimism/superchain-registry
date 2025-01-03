package config

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

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
		Addresses: Addresses{
			AddressManager:                    NewChecksummedAddress(common.Address{'A'}),
			L1CrossDomainMessengerProxy:       NewChecksummedAddress(common.Address{'B'}),
			L1ERC721BridgeProxy:               NewChecksummedAddress(common.Address{'C'}),
			L1StandardBridgeProxy:             NewChecksummedAddress(common.Address{'D'}),
			L2OutputOracleProxy:               NewChecksummedAddress(common.Address{'E'}),
			OptimismMintableERC20FactoryProxy: NewChecksummedAddress(common.Address{'F'}),
			OptimismPortalProxy:               NewChecksummedAddress(common.Address{'G'}),
			SystemConfigProxy:                 NewChecksummedAddress(common.Address{'H'}),
			ProxyAdmin:                        NewChecksummedAddress(common.Address{'I'}),
			SuperchainConfig:                  NewChecksummedAddress(common.Address{'J'}),
			AnchorStateRegistryProxy:          NewChecksummedAddress(common.Address{'K'}),
			DelayedWETHProxy:                  NewChecksummedAddress(common.Address{'L'}),
			DisputeGameFactoryProxy:           NewChecksummedAddress(common.Address{'M'}),
			FaultDisputeGame:                  NewChecksummedAddress(common.Address{'N'}),
			MIPS:                              NewChecksummedAddress(common.Address{'O'}),
			PermissionedDisputeGame:           NewChecksummedAddress(common.Address{'P'}),
			PreimageOracle:                    NewChecksummedAddress(common.Address{'Q'}),
			DAChallengeAddress:                nil,
		},
		Roles: Roles{
			SystemConfigOwner: NewChecksummedAddress(common.Address{'S'}),
			ProxyAdminOwner:   NewChecksummedAddress(common.Address{'T'}),
			Guardian:          NewChecksummedAddress(common.Address{'U'}),
			Challenger:        NewChecksummedAddress(common.Address{'V'}),
			Proposer:          nil,
			UnsafeBlockSigner: NewChecksummedAddress(common.Address{'X'}),
			BatchSubmitter:    NewChecksummedAddress(common.Address{'Y'}),
		},
	}

	expData, err := os.ReadFile("testdata/expected-addresses-with-roles.json")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, json.NewEncoder(&buf).Encode(testData))
	require.JSONEq(t, string(expData), buf.String())
}
