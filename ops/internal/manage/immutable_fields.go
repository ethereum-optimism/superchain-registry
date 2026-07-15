package manage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
)

// ConfigLifecycleViolation records a disallowed change to one chain config, found by
// comparing the working tree against a base ref.
type ConfigLifecycleViolation struct {
	// Path is the repo-relative path of the config that changed.
	Path string
	// Err describes every frozen field the change touched (from
	// [config.CheckImmutableFields]).
	Err error
}

// CheckChangedConfigLifecycles compares every chain config that changed between the
// merge-base of baseRef and HEAD (including uncommitted edits) against its committed
// version, reporting each change that violates the field-lifecycle contract
// documented in superchain/configs/README.md (see [config.CheckImmutableFields]).
//
// now decides which append-only hardfork activations are already in the past (frozen)
// versus still in the future (adjustable); callers pass the current wall-clock time.
//
// It returns the list of violations — an empty slice means every change is permitted.
// The result is advisory: callers surface it as a warning rather than a hard failure,
// so legitimate clean-ups (e.g. correcting data committed wrong) are not blocked.
func CheckChangedConfigLifecycles(root, baseRef string, now uint64) ([]ConfigLifecycleViolation, error) {
	// Compare against the merge-base (mirroring the codegen check's `git diff
	// origin/main...`) so changes that landed on the base ref after this branch forked
	// are ignored.
	mergeBase, err := gitMergeBase(root, baseRef)
	if err != nil {
		return nil, err
	}

	changed, err := changedChainConfigs(root, mergeBase)
	if err != nil {
		return nil, err
	}

	var violations []ConfigLifecycleViolation
	for _, rel := range changed {
		oldData, ok, err := gitShow(root, mergeBase, rel)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue // newly added config; no prior version to compare against
		}
		newData, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			return nil, fmt.Errorf("reading working-tree config %s: %w", rel, err)
		}

		var oldCfg, newCfg config.Chain
		if err := toml.Unmarshal(oldData, &oldCfg); err != nil {
			return nil, fmt.Errorf("unmarshaling base version of %s: %w", rel, err)
		}
		if err := toml.Unmarshal(newData, &newCfg); err != nil {
			return nil, fmt.Errorf("unmarshaling working-tree version of %s: %w", rel, err)
		}

		if err := config.CheckImmutableFields(&oldCfg, &newCfg, now); err != nil {
			violations = append(violations, ConfigLifecycleViolation{Path: rel, Err: err})
		}
	}
	return violations, nil
}

// GitRefExists reports whether a git ref can be resolved in the repo at root. Callers
// use it to skip the check when the base ref is unavailable (e.g. a shallow checkout).
func GitRefExists(root, ref string) bool {
	return exec.Command("git", "-C", root, "rev-parse", "--verify", "--quiet", ref+"^{commit}").Run() == nil
}

// gitMergeBase returns the merge-base commit of the base ref and HEAD.
func gitMergeBase(root, baseRef string) (string, error) {
	out, err := exec.Command("git", "-C", root, "merge-base", baseRef, "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("git merge-base %s HEAD: %w", baseRef, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// changedChainConfigs returns the repo-relative paths of chain config TOMLs that
// differ between the merge-base and the working tree (superchain.toml files excluded,
// matching the chain-config collector). A two-dot diff against the merge-base captures
// both committed and uncommitted changes. Deletions are excluded (--diff-filter=d):
// removing a chain is not a frozen-field change, and the file no longer exists to read.
func changedChainConfigs(root, mergeBase string) ([]string, error) {
	out, err := exec.Command("git", "-C", root, "diff", "--name-only", "--diff-filter=d", mergeBase).Output()
	if err != nil {
		return nil, fmt.Errorf("git diff against %s: %w", mergeBase, err)
	}

	var changed []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "superchain/configs/") || !strings.HasSuffix(line, ".toml") {
			continue
		}
		if filepath.Base(line) == "superchain.toml" {
			continue
		}
		changed = append(changed, line)
	}
	return changed, nil
}

// gitShow returns the contents of a repo-relative path at a given ref. The boolean is
// false when the path did not exist at that ref (e.g. a newly added config).
func gitShow(root, ref, rel string) ([]byte, bool, error) {
	out, err := exec.Command("git", "-C", root, "show", ref+":"+rel).Output()
	if err != nil {
		// git show exits non-zero when the path is absent at ref; treat as "not present"
		// rather than a hard error so newly added configs are simply skipped.
		return nil, false, nil
	}
	return out, true, nil
}
