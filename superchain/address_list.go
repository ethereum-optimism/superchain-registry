package superchain

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Roles struct {
	SystemConfigOwner Address `json:"SystemConfigOwner" toml:"SystemConfigOwner"`
	ProxyAdminOwner   Address `json:"ProxyAdminOwner" toml:"ProxyAdminOwner"`
	Guardian          Address `json:"Guardian" toml:"Guardian"`
	Challenger        Address `json:"Challenger" toml:"Challenger"`
	Proposer          Address `json:"Proposer" toml:"Proposer"`
	UnsafeBlockSigner Address `json:"UnsafeBlockSigner" toml:"UnsafeBlockSigner"`
	BatchSubmitter    Address `json:"BatchSubmitter" toml:"BatchSubmitter"`
}

// AddressList represents the set of network specific contracts and roles for a given network.
type AddressList struct {
	Roles                             `json:",inline" toml:",inline"`
	AddressManager                    Address `json:"AddressManager" toml:"AddressManager"`
	L1CrossDomainMessengerProxy       Address `json:"L1CrossDomainMessengerProxy" toml:"L1CrossDomainMessengerProxy"`
	L1ERC721BridgeProxy               Address `json:"L1ERC721BridgeProxy" toml:"L1ERC721BridgeProxy"`
	L1StandardBridgeProxy             Address `json:"L1StandardBridgeProxy" toml:"L1StandardBridgeProxy"`
	L2OutputOracleProxy               Address `json:"L2OutputOracleProxy" toml:"L2OutputOracleProxy,omitempty"`
	OptimismMintableERC20FactoryProxy Address `json:"OptimismMintableERC20FactoryProxy" toml:"OptimismMintableERC20FactoryProxy"`
	OptimismPortalProxy               Address `json:"OptimismPortalProxy,omitempty" toml:"OptimismPortalProxy,omitempty"`
	SystemConfigProxy                 Address `json:"SystemConfigProxy" toml:"SystemConfigProxy"`
	ProxyAdmin                        Address `json:"ProxyAdmin" toml:"ProxyAdmin"`
	SuperchainConfig                  Address `json:"SuperchainConfig,omitempty" toml:"SuperchainConfig,omitempty"`

	// Fault Proof contracts:
	AnchorStateRegistryProxy Address `json:"AnchorStateRegistryProxy,omitempty" toml:"AnchorStateRegistryProxy,omitempty"`
	DelayedWETHProxy         Address `json:"DelayedWETHProxy,omitempty" toml:"DelayedWETHProxy,omitempty"`
	DisputeGameFactoryProxy  Address `json:"DisputeGameFactoryProxy,omitempty" toml:"DisputeGameFactoryProxy,omitempty"`
	FaultDisputeGame         Address `json:"FaultDisputeGame,omitempty" toml:"FaultDisputeGame,omitempty"`
	MIPS                     Address `json:"MIPS,omitempty" toml:"MIPS,omitempty"`
	PermissionedDisputeGame  Address `json:"PermissionedDisputeGame,omitempty" toml:"PermissionedDisputeGame,omitempty"`
	PreimageOracle           Address `json:"PreimageOracle,omitempty" toml:"PreimageOracle,omitempty"`

	// AltDA contracts:
	DAChallengeAddress Address `json:"DAChallengeAddress,omitempty" toml:"DAChallengeAddress,omitempty"`
}

// AddressFor returns a nonzero address for the supplied name, if it has been specified
// (and an error otherwise).
func (a AddressList) AddressFor(name string) (Address, error) {
	// Use reflection to get the struct value and type
	v := reflect.ValueOf(a)

	// Try to find the field by name
	field := v.FieldByName(name)
	if !field.IsValid() {
		return Address{}, fmt.Errorf("no such name %s", name)
	}

	// Check if the field is of type Address
	if field.Type() != reflect.TypeOf(Address{}) {
		return Address{}, fmt.Errorf("field %s is not of type Address", name)
	}

	// Check if the address is a non-zero value
	address := field.Interface().(Address)
	if address == (Address{}) {
		return Address{}, fmt.Errorf("no address or zero address specified for %s", name)
	}

	return address, nil
}

// MarshalJSON excludes any addresses set to 0x000...000
func (a AddressList) MarshalJSON() ([]byte, error) {
	type AddressList2 AddressList // use another type to prevent infinite recursion later on
	b := AddressList2(a)

	o, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	out := make(map[string]Address)
	err = json.Unmarshal(o, &out)
	if err != nil {
		return nil, err
	}

	for k, v := range out {
		if (v == Address{}) {
			delete(out, k)
		}
	}

	return json.Marshal(out)
}
