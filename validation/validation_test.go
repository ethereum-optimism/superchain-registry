package validation

import (
	"testing"

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
	testUniversal(t, chain)

	if chain.SuperchainLevel == Standard ||
		(chain.SuperchainLevel == Frontier && chain.StandardChainCandidate) {
		testStandardCandidate(t, chain)
	}

	if chain.SuperchainLevel == Standard {
		testStandard(t, chain)
	}
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

// testStandardCandidate applies to Standard and Standard Candidate Chains.
func testStandardCandidate(t *testing.T, chain *ChainConfig) {
	// Standard Config Params
	t.Run("Rollup Config", func(t *testing.T) { testRollupConfig(t, chain) })
	t.Run("Resource Config", func(t *testing.T) { testResourceConfig(t, chain) })
	t.Run("Gas Limit", func(t *testing.T) { testGasLimit(t, chain) })
	t.Run("GPO Params", func(t *testing.T) { testGasPriceOracleParams(t, chain) })
	t.Run("Superchain Config", func(t *testing.T) { testSuperchainConfig(t, chain) })
	// Standard Config Roles
	t.Run("L1 Security Config", func(t *testing.T) { testL1SecurityConfig(t, chain.ChainID) })
	t.Run("L2 Security Config", func(t *testing.T) { testL2SecurityConfig(t, chain) })
	// Standard Contract Versions
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
