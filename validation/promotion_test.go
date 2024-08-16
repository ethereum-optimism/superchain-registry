package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
)

// WARNING: this test must not run along side any other tests, because it mutates global objects.
// It should be strictly isolated.
func TestPromotion(t *testing.T) {
	for _, chain := range OPChains {
		chain := chain
		t.Run(perChainTestName(chain), func(t *testing.T) {
			t.Parallel()
			if chain.StandardChainCandidate {
				// do not allow any test exclusions
				exclusions = nil
				// promote the chain to standard
				// by mutating the chainConfig
				err := chain.PromoteToStandard()
				if err != nil {
					panic(err)
				}
				testStandardCandidate(t, chain)
				testStandard(t, chain)
			}
		})
	}
}
