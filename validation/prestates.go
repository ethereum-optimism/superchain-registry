package validation

import (
	_ "embed"
	"fmt"

	"github.com/BurntSushi/toml"
)

type Prestates map[Semver]Prestate

type Prestate struct {
	ProgramVersion   string `toml:"program_version"`
	AbsolutePrestate Hash   `toml:"absolute_prestate"`
}

//go:embed standard/standard-prestates.toml
var standardPrestatesBytes []byte

var StandardPrestates Prestates

func init() {
	if err := toml.Unmarshal(standardPrestatesBytes, &StandardPrestates); err != nil {
		panic(fmt.Errorf("failed to unmarshal standard prestates: %w", err))
	}
}
