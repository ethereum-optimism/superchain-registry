package fs

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

type AtomicWriter struct {
	buf      *bytes.Buffer
	filename string
	perms    os.FileMode
}

func AtomicWrite(filename string, perms os.FileMode, data []byte) error {
	w := NewAtomicWriter(filename, perms)
	if _, err := w.Write(data); err != nil {
		return err
	}
	return w.Close()
}

func NewAtomicWriter(filename string, perms os.FileMode) *AtomicWriter {
	return &AtomicWriter{
		buf:      new(bytes.Buffer),
		filename: filename,
		perms:    perms,
	}
}

func (w *AtomicWriter) Write(data []byte) (int, error) {
	return w.buf.Write(data)
}

func (w *AtomicWriter) Close() error {
	// Stage the temp file in the destination's directory so the final rename
	// stays on the same filesystem; renaming across filesystems fails with EXDEV.
	tmp, err := os.CreateTemp(filepath.Dir(w.filename), "atomic")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	if _, err := tmp.Write(w.buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), w.filename); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	if err := os.Chmod(w.filename, w.perms); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}
	return nil
}
