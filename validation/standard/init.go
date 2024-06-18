package standard

import (
	"embed"
	"io/fs"

	"github.com/BurntSushi/toml"
)

//go:embed standard-config-params-mainnet.toml standard-config-params-sepolia.toml standard-config-params-sepolia-dev-0.toml standard-config-roles.toml
var standardConfigFile embed.FS

func init() {

	Config = ConfigType{
		Params: make(map[string]*Params),
		Roles:  new(Roles),
	}

	err := decodeTOMLFileIntoConfig("standard-config-roles.toml", Config.Roles)
	if err != nil {
		panic(err)
	}

	Config.Params["mainnet"] = new(Params)
	err = decodeTOMLFileIntoConfig("standard-config-params-mainnet.toml", Config.Params["mainnet"])
	if err != nil {
		panic(err)
	}

	Config.Params["sepolia"] = new(Params)
	err = decodeTOMLFileIntoConfig("standard-config-params-sepolia.toml", Config.Params["sepolia"])
	if err != nil {
		panic(err)
	}

	Config.Params["sepolia-dev-0"] = new(Params)
	err = decodeTOMLFileIntoConfig("standard-config-params-sepolia-dev-0.toml", Config.Params["sepolia-dev-0"])
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
