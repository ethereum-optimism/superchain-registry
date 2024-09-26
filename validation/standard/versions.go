package standard

import (
	"reflect"

	"github.com/ethereum-optimism/superchain-registry/superchain"
)

type Tag string

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
	Releases map[Tag]superchain.ContractVersions `toml:"releases"`
}

var (
	Release            Tag
	NetworkVersions                           = make(map[string]VersionTags)
	BytecodeHashes     BytecodeHashTags       = make(BytecodeHashTags, 0)
	BytecodeImmutables BytecodeImmutablesTags = make(BytecodeImmutablesTags, 0)
)

// L1ContractBytecodeHashes represents the hash of the contract bytecode (as a hex string) for each L1 contract
type L1ContractBytecodeHashes superchain.ContractBytecodeHashes

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
