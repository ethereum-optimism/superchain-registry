package genesis

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

//go:embed validation-inputs
var validationInputs embed.FS

var ValidationInputs map[uint64]ValidationMetadata

func init() {
	ValidationInputs = make(map[uint64]ValidationMetadata)

	chains, err := validationInputs.ReadDir("validation-inputs")
	if err != nil {
		panic(fmt.Errorf("failed to read validation-inputs dir: %w", err))
	}
	// iterate over superchain-target entries
	for _, s := range chains {

		if !s.IsDir() {
			continue // ignore files, e.g. a readme
		}

		// Load superchain-target config
		metadata, err := validationInputs.ReadFile(path.Join("validation-inputs", s.Name(), "meta.toml"))
		if err != nil {
			panic(fmt.Errorf("failed to read metadata file: %w", err))
		}

		m := new(ValidationMetadata)
		err = toml.Unmarshal(metadata, m)
		if err != nil {
			panic(fmt.Errorf("failed to decode metadata file: %w", err))
		}

		if strings.HasSuffix(s.Name(), "-test") {
			continue
		}

		chainID, err := strconv.Atoi(s.Name())
		if err != nil {
			panic(fmt.Errorf("failed to decode chain id from dir name: %w", err))
		}

		ValidationInputs[uint64(chainID)] = *m

		// Clone optimism into gitignored temporary directory (if that directory does not yet exist)
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			panic("No caller information")
		}
		thisDir := filepath.Dir(filename)

		if _, err := os.Stat(path.Join(thisDir, "optimism-temporary")); os.IsNotExist(err) {
			mustExecuteCommandInDir(filepath.Dir(filename),
				exec.Command("git", "clone", "https://github.com/ethereum-optimism/optimism.git", path.Join(thisDir, "optimism-temporary")))
		}
	}
}

type ValidationMetadata struct {
	GenesisCreationCommit  string `toml:"genesis_creation_commit"` // in https://github.com/ethereum-optimism/optimism/
	NodeVersion            string `toml:"node_version"`
	MonorepoBuildCommand   string `toml:"monorepo_build_command"`
	GenesisCreationCommand string `toml:"genesis_creation_command"`
}
