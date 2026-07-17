package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
)

// check_immutable_fields warns when a pull request changes a chain-config field that
// the field-lifecycle contract (superchain/configs/README.md) considers frozen: an
// immutable field, or a hardfork activation already in the past. It is advisory — the
// check-immutable-fields CI job is not a required status check, so a legitimate
// clean-up (e.g. correcting data committed wrong) surfaces a warning for a reviewer to
// acknowledge rather than being blocked.
//
// It compares the working tree against the merge-base with a base ref (origin/main by
// default, override with BASE_REF), mirroring the codegen check.
func main() {
	if err := mainErr(); err != nil {
		output.WriteNotOK("%v\n", err)
		os.Exit(1)
	}
}

func mainErr() error {
	root, err := paths.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to find repo root: %w", err)
	}

	baseRef := os.Getenv("BASE_REF")
	if baseRef == "" {
		baseRef = "origin/main"
	}
	if !manage.GitRefExists(root, baseRef) {
		output.WriteOK("base ref %q not available; skipping immutable-field check\n", baseRef)
		return nil
	}

	// now decides which hardfork activations are already in the past (frozen) versus
	// still in the future (adjustable).
	now := uint64(time.Now().Unix())

	violations, err := manage.CheckChangedConfigLifecycles(root, baseRef, now)
	if err != nil {
		return err
	}
	if len(violations) == 0 {
		output.WriteOK("no changes to frozen chain-config fields detected\n")
		return nil
	}

	for _, v := range violations {
		output.WriteWarn("%s: %v\n", v.Path, v.Err)
	}
	return fmt.Errorf("%d chain config(s) changed a frozen field; if this is an intentional correction, a reviewer should confirm it before merging", len(violations))
}
