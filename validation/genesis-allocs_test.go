package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/genesis"
	"github.com/stretchr/testify/require"
)

func testGenesisAllocsMetadata(t *testing.T, chain *ChainConfig) {
	// This tests asserts that the a genesis creation commit is stored for
	// the chain. It does not perform full genesis allocs validation.
	// Full genesis allocs validation is run as a one-off requirement when
	// chains are added to the registry and/or when the genesis creation commit itself
	// is changed (and may be re-run at other choice moments).
	// Therefore the presence of a genesis creation commit ensures that full genesis
	// validation has been performed on this chain.
	// This test is lightweight and can be run continually.
	// The test tries to ensure that the information necessary
	// for validating the genesis of the chain continues to exist over time.
	require.NotEmpty(t, genesis.ValidationInputs[chain.ChainID].GenesisCreationCommit)
}
