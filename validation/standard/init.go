package standard

import (
	"embed"
	"io/fs"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/superchain-registry/superchain"
)

//go:embed *.toml
var standardConfigFile embed.FS

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

		var versions VersionTags = VersionTags{
			Releases: make(map[Tag]superchain.ContractVersions, 0),
		}

		decodeTOMLFileIntoConfig("standard-versions-"+network+".toml", &versions)
		NetworkVersions[network] = versions
	}

	decodeTOMLFileIntoConfig("standard-bytecodes.toml", &BytecodeHashes)
	decodeTOMLFileIntoConfig("standard-immutables.toml", &BytecodeImmutables)

	// Get the single standard release Tag (universal across superchain targets)
	// and store in the standard.Release
	temp := new(struct {
		Sr string `toml:"standard_release"`
	})
	decodeTOMLFileIntoConfig("standard-releases.toml", temp)
	Release = Tag(temp.Sr)
	if Release == "" {
		panic("empty standard release")
	}
}

func decodeTOMLFileIntoConfig[
	T any](filename string, config *T) {
	data, err := fs.ReadFile(standardConfigFile, filename)
	if err != nil {
		panic(err)
	}
	err = toml.Unmarshal(data, config)
	if err != nil {
		panic(err)
	}
}
