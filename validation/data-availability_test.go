package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"
)

func TestDataAvailability(t *testing.T) {
	for _, chain := range OPChains {
		t.Run(perChainTestName(chain), func(t *testing.T) {
			RunOnStandardAndStandardCandidateChains(t, *chain)
			require.Nil(t, chain.Plasma, "Standard chains use Ethereum L1 calldata or blobs for data availability (plasma not permitted)")
		})
	}
}
