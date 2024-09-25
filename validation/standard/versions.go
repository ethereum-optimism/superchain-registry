package standard

import (
	"reflect"

	"github.com/ethereum-optimism/superchain-registry/superchain"
)

type Tag string

type (
	BytecodeHashTags       = map[Tag]L1ContractBytecodeHashes
	BytecodeImmutablesTags = map[Tag]ContractBytecodeImmutables
)

type VersionTags struct {
	Releases        map[Tag]superchain.ContractVersions `toml:"releases"`
	StandardRelease Tag                                 `toml:"standard_release,omitempty"`
}

var (
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
