package paths

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

func ReadTOMLFile(p string, out any) error {
	data, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("failed to read TOML file: %w", err)
	}

	if err := toml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("failed to unmarshal TOML: %w", err)
	}

	return nil
}

func ReadJSONFile(p string, out any) error {
	data, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}
