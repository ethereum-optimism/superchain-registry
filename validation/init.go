package validation

import (
	"embed"
	"io/fs"

	"github.com/BurntSushi/toml"
)

//go:embed standard-config-mainnet.toml standard-config-sepolia.toml
var standardConfigFile embed.FS

func init() {
	var err error
	err = decodeTOMLFileIntoConfig("standard-config-mainnet.toml", &StandardConfigMainnet)
	if err != nil {
		panic(err)
	}
	err = decodeTOMLFileIntoConfig("standard-config-sepolia.toml", &StandardConfigSepolia)
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
