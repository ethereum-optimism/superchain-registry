package validation

import (
	"testing"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
)

func TestValidation(t *testing.T) {
	// Entry point for validation checks which run
	// on each OP chain.
	for _, chain := range OPChains {
		chain := chain
		t.Run(perChainTestName(chain), func(t *testing.T) {
			t.Parallel()
			testValidation(t, chain)
		})
	}
}

func testValidation(t *testing.T, chain *ChainConfig) {
	t.Run("Universal Checks", func(t *testing.T) {
		testUniversal(t, chain)
	})

	t.Run("Standard Chain", func(t *testing.T) {
		if chain.SuperchainLevel != superchain.Standard {
			t.Skip("Chain excluded from this check (NOT a Standard Chain)")
		}
		testStandard(t, chain)
	})

	t.Run("Standard or Standard Candidate Chain", func(t *testing.T) {
		if !chain.StandardChainCandidate && chain.SuperchainLevel != Standard {
			t.Skip("Chain excluded from this check (NOT a Standard or a Standard Candidate Chain)")
		}
		testStandardCandidate(t, chain)
	})
}

// testUniversal should be applied to each chain in the registry
// regardless of superchain_level. There should be no
// exceptions, since the test is e.g.
// designed to protect downstream software or
// sanity checking basic consistency conditions.
func testUniversal(t *testing.T, chain *ChainConfig) {
	t.Run("Genesis Hash Check", func(t *testing.T) {
		testGenesisHash(t, chain.ChainID)
	})
	t.Run("Genesis RPC Check", func(t *testing.T) {
		testGenesisHashAgainstRPC(t, chain)
	})
	t.Run("Uniqueness Check", func(t *testing.T) {
		testIsGloballyUnique(t, chain)
	})
	t.Run("ChainID RPC Check", func(t *testing.T) {
		testChainIDFromRPC(t, chain)
	})
}

// testStandardCandidate should be applied only to a fully Standard Chain,
// i.e. not to a Standard Candidate Chain.
func testStandardCandidate(t *testing.T, chain *ChainConfig) {
	t.Run("Standard Config Params", func(t *testing.T) {
		t.Run("Data Availabilty", func(t *testing.T) { testDataAvailability(t, chain) })
		t.Run("Resource Config", func(t *testing.T) { testResourceConfig(t, chain) })
		t.Run("L2OO Params", func(t *testing.T) { testL2OOParams(t, chain) })
		t.Run("Gas Limit", func(t *testing.T) { testGasLimit(t, chain) })
		t.Run("GPO Params", func(t *testing.T) { testGasPriceOracleParams(t, chain) })
	})
	t.Run("Standard Config Roles", func(t *testing.T) {
		t.Run("L1 Security Config", func(t *testing.T) { testL1SecurityConfig(t, chain.ChainID) })
		t.Run("L2 Security Config", func(t *testing.T) { testL2SecurityConfig(t, chain) })
	})
	t.Run("Standard Contract Versions", func(t *testing.T) {
		testContractsMatchATag(t, chain)
	})
}

// testStandard should be applied only to a fully Standard Chain,
// i.e. not to a Standard Candidate Chain.
func testStandard(t *testing.T, chain *ChainConfig) {
	t.Run("Key Handover Check", func(t *testing.T) {
		testKeyHandover(t, chain.ChainID)
	})
}
