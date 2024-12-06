package standard

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/superchain-registry/superchain"
)

//go:embed config
var standardConfigFS embed.FS

func init() {
	Config = ConfigType{
		Params:        make(map[string]*Params),
		Roles:         new(Roles),
		MultisigRoles: make(map[string]*MultisigRoles),
	}

	decodeTOMLFileIntoConfig("standard-config-roles-universal.toml", Config.Roles)

	networkDirName := "networks"
	networks, err := standardConfigFS.ReadDir("config/" + networkDirName)
	if err != nil {
		panic(fmt.Errorf("failed to read dir: %w", err))
	}

	// iterate over network entries
	for _, networkDir := range networks {
		if !networkDir.IsDir() {
			continue // ignore files, e.g. a readme
		}

		network := networkDir.Name()

		Config.MultisigRoles[network] = new(MultisigRoles)
		decodeTOMLFileIntoConfig(networkDirName+"/"+network+"/standard-config-roles-"+network+".toml", Config.MultisigRoles[network])

		Config.Params[network] = new(Params)
		decodeTOMLFileIntoConfig(networkDirName+"/"+network+"/standard-config-params-"+network+".toml", Config.Params[network])

		var versions VersionTags = VersionTags{
			Releases: make(map[Tag]superchain.ContractVersions, 0),
		}

		decodeTOMLFileIntoConfig(networkDirName+"/"+network+"/standard-versions-"+network+".toml", &versions)
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
	data, err := fs.ReadFile(standardConfigFS, "config/"+filename)
	if err != nil {
		panic(err)
	}
	err = toml.Unmarshal(data, config)
	if err != nil {
		panic(err)
	}
}
