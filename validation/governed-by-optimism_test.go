package validation

import (
	"fmt"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"

	"github.com/stretchr/testify/require"
)

func testGovernedByOptimism(t *testing.T, chain *ChainConfig) {
	chainID := chain.ChainID
	superchain := OPChains[chainID].Superchain

	superchainRoleConfig := standard.Config.MultisigRoles[superchain]

	if superchainRoleConfig == nil {
		t.Errorf("No role configuration found for superchain '%s'!", superchain)
		return
	}
	
	optimismMultisig := superchainRoleConfig.KeyHandover.L1.Universal["ProxyAdmin"]["owner()"]

	makeMsg := func(governed bool) string {
		var notString string
		if !governed {
			notString = "not "
		}
		return fmt.Sprintf("Chains %susing Optimism governance must %shave their ProxyAdminOwner set to the Optimism multisig", notString, notString)
	}

	paoAddress := chain.Addresses.ProxyAdminOwner.String()
	
	if chain.GovernedByOptimism == true {	
		require.Equal(t, optimismMultisig, paoAddress, makeMsg(true))
	} else {
		require.NotEqual(t, optimismMultisig, paoAddress, makeMsg(false))
	}
}
