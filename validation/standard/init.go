package standard

import (
	"embed"
	"io/fs"

	"github.com/BurntSushi/toml"
)

//go:embed standard-config-params-mainnet.toml standard-config-params-sepolia.toml standard-config-params-sepolia-dev-0.toml standard-config-roles.toml
var standardConfigFile embed.FS

func init() {

	Config = make(map[string]ConfigType)

	Config["mainnet"] = ConfigType{new(Params), new(Roles)}
	var err error
	err = decodeTOMLFileIntoConfig("standard-config-params-mainnet.toml", Config["mainnet"].Params)
	if err != nil {
		panic(err)
	}
	err = decodeTOMLFileIntoConfig("standard-config-roles.toml", Config["mainnet"].Roles)
	if err != nil {
		panic(err)
	}

	Config["sepolia"] = ConfigType{new(Params), new(Roles)}
	err = decodeTOMLFileIntoConfig("standard-config-params-sepolia.toml", Config["sepolia"].Params)
	if err != nil {
		panic(err)
	}
	err = decodeTOMLFileIntoConfig("standard-config-roles.toml", Config["sepolia"].Roles)
	if err != nil {
		panic(err)
	}

	Config["sepolia-dev-0"] = ConfigType{new(Params), new(Roles)}
	err = decodeTOMLFileIntoConfig("standard-config-params-sepolia-dev-0.toml", Config["sepolia-dev-0"].Params)
	if err != nil {
		panic(err)
	}
	err = decodeTOMLFileIntoConfig("standard-config-roles.toml", Config["sepolia-dev-0"].Roles)
	if err != nil {
		panic(err)
	}
}

func decodeTOMLFileIntoConfig[T Params | Roles](filename string, config *T) error {
	data, err := fs.ReadFile(standardConfigFile, filename)
	if err != nil {
		return err
	}
	return toml.Unmarshal(data, config)
}
