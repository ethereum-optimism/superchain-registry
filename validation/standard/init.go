package standard

import (
	"embed"
	"io/fs"

	"github.com/BurntSushi/toml"
)

//go:embed standard-config-mainnet.toml standard-config-sepolia.toml standard-config-sepolia-dev-0.toml
var standardConfigFile embed.FS

func init() {
	Config = make(map[string]*ConfigType)

	Config["mainnet"] = new(ConfigType)
	var err error
	err = decodeTOMLFileIntoConfig("standard-config-mainnet.toml", Config["mainnet"])
	if err != nil {
		panic(err)
	}

	Config["sepolia"] = new(ConfigType)
	err = decodeTOMLFileIntoConfig("standard-config-sepolia.toml", Config["sepolia"])
	if err != nil {
		panic(err)
	}

	Config["sepolia-dev-0"] = new(ConfigType)
	err = decodeTOMLFileIntoConfig("standard-config-sepolia-dev-0.toml", Config["sepolia-dev-0"])
	if err != nil {
		panic(err)
	}
}

func decodeTOMLFileIntoConfig(filename string, config *ConfigType) error {
	data, err := fs.ReadFile(standardConfigFile, filename)
	if err != nil {
		return err
	}
	return toml.Unmarshal(data, config)
}
