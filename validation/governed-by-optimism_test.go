package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"

	// "github.com/ethereum/go-ethereum/common"

	"github.com/stretchr/testify/require"
)

func testGovernedByOptimism(t *testing.T, chain *ChainConfig) {
	chainID := chain.ChainID
	superchain := OPChains[chainID].Superchain

	if chain.GovernedByOptimism == true {
		optimismMultisig := standard.Config.MultisigRoles[superchain].KeyHandover.L1.Universal["ProxyAdmin"]["owner()"]
		require.Equal(t, optimismMultisig, chain.Addresses.ProxyAdminOwner.String(), "Chains using Optimism governance must have their ProxyAdminOwner set to the Optimism multisig")
	}
}
