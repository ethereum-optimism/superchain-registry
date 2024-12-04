package standard

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"golang.org/x/mod/semver"
)

type Tag string

type MappedContractProperties[T string | VersionedContract] struct {
	L1CrossDomainMessenger       T `toml:"l1_cross_domain_messenger,omitempty"`
	L1ERC721Bridge               T `toml:"l1_erc721_bridge,omitempty"`
	L1StandardBridge             T `toml:"l1_standard_bridge,omitempty"`
	L2OutputOracle               T `toml:"l2_output_oracle,omitempty"`
	OptimismMintableERC20Factory T `toml:"optimism_mintable_erc20_factory,omitempty"`
	OptimismPortal               T `toml:"optimism_portal,omitempty"`
	OptimismPortal2              T `toml:"optimism_portal2,omitempty"`
	SystemConfig                 T `toml:"system_config,omitempty"`
	// Superchain-wide contracts:
	ProtocolVersions T `toml:"protocol_versions,omitempty"`
	SuperchainConfig T `toml:"superchain_config,omitempty"`
	// Fault Proof contracts:
	AnchorStateRegistry     T `toml:"anchor_state_registry,omitempty"`
	DelayedWETH             T `toml:"delayed_weth,omitempty"`
	DisputeGameFactory      T `toml:"dispute_game_factory,omitempty"`
	FaultDisputeGame        T `toml:"fault_dispute_game,omitempty"`
	MIPS                    T `toml:"mips,omitempty"`
	PermissionedDisputeGame T `toml:"permissioned_dispute_game,omitempty"`
	PreimageOracle          T `toml:"preimage_oracle,omitempty"`
	CannonFaultDisputeGame  T `toml:"cannon_fault_dispute_game,omitempty"`
}

// ContractBytecodeHashes stores a bytecode hash against each contract
type ContractBytecodeHashes MappedContractProperties[string]

// VersionedContract represents a contract that has a semantic version.
type VersionedContract struct {
	Version string `toml:"version"`
	// If the contract is a superchain singleton, it will have a static address
	Address *superchain.Address `toml:"implementation_address,omitempty"`
	// If the contract is proxied, the implementation will have a static address
	ImplementationAddress *superchain.Address `toml:"address,omitempty"`
}

// ContractVersions represents the desired semantic version of the contracts
// in the superchain. This currently only supports L1 contracts but could
// represent L2 predeploys in the future.
type ContractVersions MappedContractProperties[VersionedContract]

// GetNonEmpty returns a slice of contract names, with an entry for each contract
// in the receiver with a non empty Version property.
func (c ContractVersions) GetNonEmpty() []string {
	// Get the value and type of the struct
	v := reflect.ValueOf(c)
	t := reflect.TypeOf(c)

	var fieldNames []string

	// Iterate through the struct fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Ensure the field is of type VersionedContract
		if field.Type() == reflect.TypeOf(VersionedContract{}) {
			// Get the Version field from the VersionedContract
			versionField := field.FieldByName("Version")

			// Check if the Version is non-empty
			if versionField.IsValid() && versionField.String() != "" {
				fieldNames = append(fieldNames, fieldType.Name)
			}
		}
	}

	return fieldNames
}

// VersionFor returns the version for the supplied contract name, if it exits
// (and an error otherwise). Useful for slicing into the struct using a string.
func (c ContractVersions) VersionFor(contractName string) (string, error) {
	// Use reflection to get the value of the struct
	val := reflect.ValueOf(c)
	// Get the field by name (contractName)
	field := val.FieldByName(contractName)

	// Check if the field exists and is a struct
	if !field.IsValid() {
		return "", errors.New("no such contract name")
	}

	// Check if the struct contains the "Version" field
	versionField := field.FieldByName("Version")
	if !versionField.IsValid() || versionField.String() == "" {
		return "", errors.New("no version specified")
	}

	// Return the version if it's a string
	if versionField.Kind() == reflect.String {
		return versionField.String(), nil
	}

	return "", errors.New("version is not a string")
}

// Check will sanity check the validity of the semantic version strings
// in the ContractVersions struct. If allowEmptyVersions is true, empty version errors will be ignored.
func (c ContractVersions) Check(allowEmptyVersions bool) error {
	val := reflect.ValueOf(c)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		vC, ok := field.Interface().(VersionedContract)
		if !ok {
			return fmt.Errorf("invalid type for field %s", val.Type().Field(i).Name)
		}
		if vC.Version == "" {
			if allowEmptyVersions {
				continue // we allow empty strings and rely on tests to assert (or except) a nonempty version
			}
			return fmt.Errorf("empty version for field %s", val.Type().Field(i).Name)
		}
		vC.Version = CanonicalizeSemver(vC.Version)
		if !semver.IsValid(vC.Version) {
			return fmt.Errorf("invalid semver %s for field %s", vC.Version, val.Type().Field(i).Name)
		}
	}
	return nil
}

// ContractBytecodeImmutables stores the immutable references as a raw stringified JSON string in a TOML config.
// it is stored this way because it can be plucked out of the contract compilation output as is and pasted into the TOML config file.
type ContractBytecodeImmutables struct {
	AnchorStateRegistry string `toml:"anchor_state_registry,omitempty"`
	DelayedWETH         string `toml:"delayed_weth,omitempty"`
	FaultDisputeGame    string `toml:"fault_dispute_game,omitempty"`
	MIPS                string `toml:"mips,omitempty"`
}

func (c ContractBytecodeImmutables) ForContractWithName(name string) (string, bool) {
	// Use reflection to get the struct value and type
	v := reflect.ValueOf(c)

	// Try to find the field by name
	field := v.FieldByName(name)
	if !field.IsValid() {
		return "", false
	}

	// Check if the field is of type string
	if field.Type() != reflect.TypeOf("") {
		return "", false
	}

	// Check if the string is empty
	s := field.Interface().(string)
	if s == "" {
		return "", false
	}

	return s, true
}

type (
	BytecodeHashTags       = map[Tag]L1ContractBytecodeHashes
	BytecodeImmutablesTags = map[Tag]ContractBytecodeImmutables
)

type VersionTags struct {
	Releases map[Tag]ContractVersions `toml:"releases"`
}

var (
	Release            Tag
	NetworkVersions    = make(map[string]VersionTags)
	BytecodeHashes     = make(BytecodeHashTags, 0)
	BytecodeImmutables = make(BytecodeImmutablesTags, 0)
)

// L1ContractBytecodeHashes represents the hash of the contract bytecode (as a hex string) for each L1 contract
type L1ContractBytecodeHashes ContractBytecodeHashes

// GetNonEmpty returns a slice of contract name strings, with an entry for each key in the receiver
// with a non-empty value
func (bch L1ContractBytecodeHashes) GetNonEmpty() []string {
	// Get the value and type of the struct
	v := reflect.ValueOf(bch)
	t := reflect.TypeOf(bch)

	var fieldNames []string

	// Iterate through the struct fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Ensure the field is of type string
		if field.Kind() == reflect.String {
			// Check if the string field is non-empty
			if field.String() != "" {
				fieldNames = append(fieldNames, fieldType.Name)
			}
		}
	}

	return fieldNames
}

// CanonicalizeSemver will ensure that the version string has a "v" prefix.
// This is because the semver library being used requires the "v" prefix,
// even though
func CanonicalizeSemver(version string) string {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return version
}
