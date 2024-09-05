package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/common"
)

// WARNING: this test must not run along side any other tests, because it mutates global objects.
// It should be strictly isolated.
func TestPromotion(t *testing.T) {
	for _, chain := range OPChains {
		chain := chain
		t.Run(common.PerChainTestName(chain), func(t *testing.T) {
			t.Parallel()
			if chain.StandardChainCandidate {
				// do not allow any test exclusions
				exclusions = nil
				// simulate promoting the chain to standard
				// by mutating a copy of the chainConfig
				copy, err := chain.PromoteToStandard()
				if err != nil {
					panic(err)
				}
				testStandardCandidate(t, copy)
				testStandard(t, copy)
			}
		})
	}
}
