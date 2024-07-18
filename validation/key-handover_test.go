package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
)

func testKeyHandover(t *testing.T, chain *ChainConfig) {
	client := clients.L1[chain.Superchain]

	// L1 Proxy Admin
	checkResolutions(t, standard.Config.MultisigRoles[chain.Superchain].KeyHandover.L1.Universal, chain.ChainID, client)

	client = clients.L2[chain.ChainID]

	// L2 Proxy Admin
	checkResolutions(t, standard.Config.MultisigRoles[chain.Superchain].KeyHandover.L2.Universal, chain.ChainID, client)
}
