package standard

import (
	"embed"
	"io/fs"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/superchain-registry/superchain"
)

//go:embed *.toml
var standardConfigFile embed.FS

// ContractASTsWithImmutableReferences caches the `immutableReferences` after parsing it from the config file.
//
//	The config file is generated from the compiled contract AST (from the combined JSON
//
// artifact from the monorepo. We do this because the contracts and compiled artifacts are not available in the superchain
// registry. Ex: ethereum-optimism/optimism/packages/contracts-bedrock/forge-artifacts/MIPS.sol/MIPS.json
var ContractASTsWithImmutableReferences = map[string]string{}

// L1ContractBytecodeHashes represents the hash of the contract bytecode (as a hex string) for each L1 contract
type L1ContractBytecodeHashes superchain.ContractVersions

// ContractBytecodeImmutables stores the immutable references as a raw stringified JSON string in a TOML config.
// it is stored this way because it can be plucked out of the contract compilation output as is and pasted into the TOML config file.
type ContractBytecodeImmutables struct {
	AnchorStateRegistry string `toml:"anchor_state_registry,omitempty"`
	DelayedWETH         string `toml:"delayed_weth,omitempty"`
	FaultDisputeGame    string `toml:"fault_dispute_game,omitempty"`
	MIPS                string `toml:"mips,omitempty"`
}

func init() {
	Config = ConfigType{
		Params:        make(map[string]*Params),
		Roles:         new(Roles),
		MultisigRoles: make(map[string]*MultisigRoles),
	}

	decodeTOMLFileIntoConfig("standard-config-roles-universal.toml", Config.Roles)

	networks := []string{"mainnet", "sepolia"}
	for _, network := range networks {
		Config.MultisigRoles[network] = new(MultisigRoles)
		decodeTOMLFileIntoConfig("standard-config-roles-"+network+".toml", Config.MultisigRoles[network])

		Config.Params[network] = new(Params)
		decodeTOMLFileIntoConfig("standard-config-params-"+network+".toml", Config.Params[network])
	}

	decodeTOMLFileIntoConfig("standard-versions.toml", &Versions)
	decodeTOMLFileIntoConfig("standard-bytecodes.toml", &BytecodeHashes)
	decodeTOMLFileIntoConfig("standard-immutables.toml", &BytecodeImmutables)

	LoadImmutableReferences()
}

func decodeTOMLFileIntoConfig[T Params | Roles | MultisigRoles | VersionTags | BytecodeHashTags | BytecodeImmutablesTags](filename string, config *T) {
	data, err := fs.ReadFile(standardConfigFile, filename)
	if err != nil {
		panic(err)
	}
	err = toml.Unmarshal(data, config)
	if err != nil {
		panic(err)
	}
}

// LoadImmutableReferences parses standard-immutables.toml and stores it in a map. Needs to be invoked one-time only.
func LoadImmutableReferences() {
	var bytecodeImmutables *ContractBytecodeImmutables
	for tag := range Versions {
		for contractVersion, immutables := range BytecodeImmutables {
			if tag == contractVersion {
				bytecodeImmutables = &immutables
				break
			}
		}
	}
	if bytecodeImmutables != nil {
		s := reflect.ValueOf(bytecodeImmutables).Elem()
		for i := 0; i < s.NumField(); i++ {
			name := s.Type().Field(i).Name
			value := string(reflect.ValueOf(*bytecodeImmutables).Field(i).String())
			ContractASTsWithImmutableReferences[name] = value
		}
	}
}
