package config

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/optimism/op-chain-ops/addresses"
	"github.com/ethereum-optimism/optimism/op-fetcher/pkg/fetcher/fetch/script"
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
			SuperFaultDisputeGame:             NewChecksummedAddress(common.Address{'R'}),
			SuperPermissionedDisputeGame:      NewChecksummedAddress(common.Address{'Z'}),
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

func TestCreateAddressesWithRolesFromFetcher(t *testing.T) {
	legacyFDG := common.Address{'N'}
	legacyPDG := common.Address{'P'}
	superFDG := common.Address{'S', 'F'}
	superPDG := common.Address{'S', 'P'}

	// Every role and every other address is left zero, and MarshalJSON drops zero
	// addresses, so the marshalled map holds exactly the game impls set below.
	tests := []struct {
		name   string
		addrs  script.Addresses
		expect map[string]string
	}{
		{
			name: "legacy cannon game types",
			addrs: script.Addresses{
				OpChainContracts: addresses.OpChainContracts{
					OpChainFaultProofsContracts: addresses.OpChainFaultProofsContracts{
						FaultDisputeGameImpl:        legacyFDG,
						PermissionedDisputeGameImpl: legacyPDG,
					},
				},
			},
			expect: map[string]string{
				"FaultDisputeGame":        NewChecksummedAddress(legacyFDG).String(),
				"PermissionedDisputeGame": NewChecksummedAddress(legacyPDG).String(),
			},
		},
		{
			// A chain that has migrated to super-root games registers only
			// SUPER_CANNON_KONA and SUPER_PERMISSIONED, so the legacy impls read back zero.
			name: "super-root game types only",
			addrs: script.Addresses{
				OpChainContracts: addresses.OpChainContracts{
					OpChainFaultProofsContracts: addresses.OpChainFaultProofsContracts{
						SuperFaultDisputeGameImpl:        superFDG,
						SuperPermissionedDisputeGameImpl: superPDG,
					},
				},
			},
			expect: map[string]string{
				"SuperFaultDisputeGame":        NewChecksummedAddress(superFDG).String(),
				"SuperPermissionedDisputeGame": NewChecksummedAddress(superPDG).String(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := json.Marshal(CreateAddressesWithRolesFromFetcher(test.addrs, addresses.OpChainRoles{}))
			require.NoError(t, err)

			var got map[string]string
			require.NoError(t, json.Unmarshal(data, &got))

			require.Equal(t, test.expect, got)
		})
	}
}
