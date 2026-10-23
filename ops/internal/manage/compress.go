package manage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/fs"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/klauspost/compress/zstd"
)

// WriteSuperchainGenesis writes a new chain's genesis JSON into the superchain's genesis directory.
func WriteSuperchainGenesis(rootP string, superchain config.Superchain, shortName string, gen json.RawMessage) error {
	genPath := paths.GenesisFile(rootP, superchain, shortName)
	exists, err := fs.FileExists(genPath)
	if err != nil {
		return fmt.Errorf("failed to check if genesis exists: %w", err)
	}
	if exists {
		return fmt.Errorf("genesis already exists: %s", genPath)
	}

	if err := os.MkdirAll(path.Dir(genPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return WriteGenesis(rootP, genPath, gen)
}

// WriteGenesis compresses the genesis JSON gen with the registry's dictionary and writes it to
// genPath. gen is stored byte for byte rather than decoded into a genesis type, which would drop
// any field that type doesn't know.
func WriteGenesis(rootP string, genPath string, gen json.RawMessage) error {
	if !json.Valid(gen) {
		return errors.New("genesis is not valid JSON")
	}

	dictPath := path.Join(paths.ExtraDir(rootP), "dictionary")
	dict, err := os.ReadFile(dictPath)
	if err != nil {
		return fmt.Errorf("failed to read dictionary: %w", err)
	}

	buf := new(bytes.Buffer)
	zr, err := zstd.NewWriter(buf, zstd.WithEncoderDict(dict))
	if err != nil {
		return fmt.Errorf("failed to create zstd writer: %w", err)
	}
	if _, err := zr.Write(gen); err != nil {
		return fmt.Errorf("failed to compress genesis: %w", err)
	}
	if err := zr.Close(); err != nil {
		return fmt.Errorf("failed to close zstd writer: %w", err)
	}

	if err := fs.AtomicWrite(genPath, 0o755, buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write genesis: %w", err)
	}

	return nil
}

// ReadSuperchainGenesis reads a chain's genesis JSON from the superchain's genesis directory.
func ReadSuperchainGenesis(rootP string, superchain config.Superchain, shortName string) (json.RawMessage, error) {
	genPath := paths.GenesisFile(rootP, superchain, shortName)
	exists, err := fs.FileExists(genPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if genesis exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("genesis does not exist: %s", genPath)
	}

	return ReadGenesis(rootP, genPath)
}

// ReadGenesis decompresses the genesis at genPath and returns its JSON byte for byte.
func ReadGenesis(rootP string, genPath string) (json.RawMessage, error) {
	dictPath := path.Join(paths.ExtraDir(rootP), "dictionary")
	dict, err := os.ReadFile(dictPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dictionary: %w", err)
	}

	genF, err := os.Open(genPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open genesis: %w", err)
	}
	defer genF.Close()

	zr, err := zstd.NewReader(genF, zstd.WithDecoderDicts(dict))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd reader: %w", err)
	}
	defer zr.Close()

	gen, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress genesis: %w", err)
	}
	if !json.Valid(gen) {
		return nil, errors.New("genesis is not valid JSON")
	}

	return gen, nil
}
