package standard

import (
	"embed"
	"io/fs"

	"github.com/BurntSushi/toml"
)

//go:embed *.toml
var standardConfigFile embed.FS

func init() {
	Config = ConfigType{
		Params:        make(map[string]*Params),
		Roles:         new(Roles),
		MultisigRoles: make(map[string]*MultisigRoles),
	}

	Versions = make(map[string]*VersionTags)

	decodeTOMLFileIntoConfig("standard-config-roles-universal.toml", Config.Roles)

	networks := []string{"mainnet", "sepolia"}
	for _, network := range networks {
		Config.MultisigRoles[network] = new(MultisigRoles)
		decodeTOMLFileIntoConfig("standard-config-roles-"+network+".toml", Config.MultisigRoles[network])

		Config.Params[network] = new(Params)
		decodeTOMLFileIntoConfig("standard-config-params-"+network+".toml", Config.Params[network])

		Versions[network] = new(VersionTags)
		decodeTOMLFileIntoConfig("standard-versions-"+network+".toml", Versions[network])
	}

}

func decodeTOMLFileIntoConfig[T Params | Roles | MultisigRoles | VersionTags](filename string, config *T) {
	data, err := fs.ReadFile(standardConfigFile, filename)
	if err != nil {
		panic(err)
	}
	err = toml.Unmarshal(data, config)
	if err != nil {
		panic(err)
	}
}
