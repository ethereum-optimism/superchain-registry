package validation

import (
	"testing"
	"time"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
)

func TestPromotion(t *testing.T) {
	for _, chain := range OPChains {
		chain := chain
// WARNING: this test must not run along side any other tests, because it mutates some global object.
// It should be strictly isolated. 
		t.Run(perChainTestName(chain), func(t *testing.T) {
			t.Parallel()
			if chain.StandardChainCandidate {
				// do not allow any test exclusions
				exclusions = nil
				// promote the chain to standard
				// by mutating the chainConfig
				chain.StandardChainCandidate = false
				chain.SuperchainLevel = Standard
				now := uint64(time.Now().Unix())
				chain.SuperchainTime = &now
				testStandardCandidate(t, chain)
				testStandard(t, chain)
			}
		})
	}
}
