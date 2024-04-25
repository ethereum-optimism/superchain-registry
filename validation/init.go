package validation

import (
	"embed"
	"io/fs"

	"github.com/BurntSushi/toml"
)

//go:embed standard-config-mainnet.toml standard-config-sepolia.toml standard-config-sepolia-dev-0.toml
var standardConfigFile embed.FS

func init() {
	StandardConfig = make(map[string]*StandardConfigTy)

	StandardConfig["mainnet"] = new(StandardConfigTy)
	var err error
	err = decodeTOMLFileIntoConfig("standard-config-mainnet.toml", StandardConfig["mainnet"])
	if err != nil {
		panic(err)
	}

	StandardConfig["sepolia"] = new(StandardConfigTy)
	err = decodeTOMLFileIntoConfig("standard-config-sepolia.toml", StandardConfig["sepolia"])
	if err != nil {
		panic(err)
	}

	StandardConfig["sepolia-dev-0"] = new(StandardConfigTy)
	err = decodeTOMLFileIntoConfig("standard-config-sepolia-dev-0.toml", StandardConfig["sepolia-dev-0"])
	if err != nil {
		panic(err)
	}
}

func decodeTOMLFileIntoConfig(filename string, config *StandardConfigTy) error {
	data, err := fs.ReadFile(standardConfigFile, filename)
	if err != nil {
		return err
	}
	return toml.Unmarshal(data, config)
}
