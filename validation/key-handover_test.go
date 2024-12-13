package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/ethereum/go-ethereum/ethclient"
)

func testKeyHandover(t *testing.T, chain *ChainConfig) {
	chainID := chain.ChainID
	superchain := OPChains[chainID].Superchain
	
	rpcEndpoint := Superchains[superchain].Config.L1.PublicRPC
	client, _ := ethclient.Dial(rpcEndpoint)

	// L1 Proxy Admin
	checkResolutions(t, standard.Config.MultisigRoles[superchain].KeyHandover.L1.Universal, chainID, client)

	rpcEndpoint = OPChains[chainID].PublicRPC
	client, _ = ethclient.Dial(rpcEndpoint)

	// L2 Proxy Admin
	checkResolutions(t, standard.Config.MultisigRoles[superchain].KeyHandover.L2.Universal, chainID, client)
}
