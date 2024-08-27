package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

func LoadJSON[X any](inputPath string) (*X, error) {
	if inputPath == "" {
		return nil, errors.New("no path specified")
	}
	f, err := os.OpenFile(inputPath, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", inputPath, err)
	}
	defer f.Close()
	var obj X
	if err := json.NewDecoder(f).Decode(&obj); err != nil {
		return nil, fmt.Errorf("failed to decode file %q: %w", inputPath, err)
	}
	return &obj, nil
}
