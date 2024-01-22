package superchain_test

import (
	"testing"

	"github.com/ethereum-optimism/superchain-registry/superchain"
)

// TestContractVersionsCheck will fail if the superchain semver file
// is not read correctly.
func TestContractVersionsCheck(t *testing.T) {
	if err := superchain.SuperchainSemver.Check(); err != nil {
		t.Fatal(err)
	}
}
