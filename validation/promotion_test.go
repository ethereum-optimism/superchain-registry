package validation

import (
	"testing"
	"time"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
)

func TestPromotion(t *testing.T) {
	for _, chain := range OPChains {
		chain := chain
		t.Run(perChainTestName(chain), func(t *testing.T) {
			t.Parallel()
			if chain.StandardChainCandidate {
				chain.StandardChainCandidate = false
				chain.SuperchainLevel = Standard
				now := uint64(time.Now().Unix())
				chain.SuperchainTime = &now
				testStandard(t, chain)
			}
		})
	}
}
