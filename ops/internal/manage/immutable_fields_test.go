package manage

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/stretchr/testify/require"
)

// TestImmutableFieldsUnchanged enforces the field-lifecycle contract documented in
// superchain/configs/README.md: immutable fields (e.g. chain_id, genesis,
// block_time) may never change once a chain is registered, and hardfork activations
// are append-only. It compares every chain config changed on this branch against its
// committed version on the base ref and fails if a disallowed change slipped in.
//
// The base ref defaults to origin/main (matching the CI codegen check) and can be
// overridden with the BASE_REF env var. When the base ref cannot be resolved (e.g. a
// shallow checkout with no remote), the test skips rather than failing spuriously.
func TestImmutableFieldsUnchanged(t *testing.T) {
	root, err := paths.FindRepoRoot()
	require.NoError(t, err)

	baseRef := os.Getenv("BASE_REF")
	if baseRef == "" {
		baseRef = "origin/main"
	}
	if !gitRefExists(t, root, baseRef) {
		t.Skipf("base ref %q not available; skipping immutable-field check", baseRef)
	}

	// Compare against the merge-base (mirroring CI's `git diff origin/main...`) so
	// changes that landed on the base ref since this branch forked are ignored.
	mergeBase := gitMergeBase(t, root, baseRef)

	changed := changedChainConfigs(t, root, mergeBase)
	if len(changed) == 0 {
		t.Log("no superchain/configs/*.toml files changed against base ref")
		return
	}

	// now decides which hardfork activations are already in the past (frozen) vs.
	// still in the future (adjustable).
	now := uint64(time.Now().Unix())

	for _, rel := range changed {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			oldData, ok := gitShow(t, root, mergeBase, rel)
			if !ok {
				t.Logf("%s is newly added; no prior version to compare against", rel)
				return
			}
			newData, err := os.ReadFile(filepath.Join(root, rel))
			require.NoError(t, err, "reading working-tree config")

			var oldCfg, newCfg config.Chain
			require.NoError(t, toml.Unmarshal(oldData, &oldCfg), "unmarshaling base version")
			require.NoError(t, toml.Unmarshal(newData, &newCfg), "unmarshaling working-tree version")

			require.NoError(t, config.CheckImmutableFields(&oldCfg, &newCfg, now))
		})
	}
}

// gitRefExists reports whether a git ref can be resolved in the repo at root.
func gitRefExists(t *testing.T, root, ref string) bool {
	t.Helper()
	cmd := exec.Command("git", "-C", root, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	return cmd.Run() == nil
}

// gitMergeBase returns the merge-base commit of the base ref and HEAD.
func gitMergeBase(t *testing.T, root, baseRef string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", root, "merge-base", baseRef, "HEAD")
	out, err := cmd.Output()
	require.NoError(t, err, "git merge-base")
	return strings.TrimSpace(string(out))
}

// changedChainConfigs returns the repo-relative paths of chain config TOMLs that
// differ between the merge-base and the working tree (superchain.toml files
// excluded, matching the chain-config collector). A two-dot diff against the
// merge-base captures both committed and uncommitted changes. Deletions are
// excluded (--diff-filter=d): removing a chain is not an immutable-field change,
// and the file no longer exists to read.
func changedChainConfigs(t *testing.T, root, mergeBase string) []string {
	t.Helper()
	cmd := exec.Command("git", "-C", root, "diff", "--name-only", "--diff-filter=d", mergeBase)
	out, err := cmd.Output()
	require.NoError(t, err, "git diff against merge-base")

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
	return changed
}

// gitShow returns the contents of a repo-relative path at a given ref. The boolean
// is false when the path did not exist at that ref (e.g. a newly added config).
func gitShow(t *testing.T, root, ref, rel string) ([]byte, bool) {
	t.Helper()
	cmd := exec.Command("git", "-C", root, "show", ref+":"+rel)
	out, err := cmd.Output()
	if err != nil {
		return nil, false
	}
	return out, true
}
