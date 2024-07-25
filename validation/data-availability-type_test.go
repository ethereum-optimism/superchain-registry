package validation

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/superchain-registry/superchain"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func testDataAvailabilityType(t *testing.T, chain *ChainConfig) {
	rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint)

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	// Standard chains must be eth-da for now. Will add checks for alt-da later
	require.Equal(t, superchain.EthDA, chain.DataAvailabilityType)

	batchSubmitterAddress := Addresses[chain.ChainID].BatchSubmitter
	require.NotZero(t, batchSubmitterAddress)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	depth := chain.SequencerWindowSize
	require.NotZero(t, depth)

	// First attempt without needing an archive node
	//  - full nodes keep the last 128 blocks but we use 120 to give a buffer
	nonArchivalDepth := 120
	_, found, err := eth.CheckRecentTxs(ctx, client, nonArchivalDepth, common.Address(batchSubmitterAddress))
	if !found {
		_, found, err = eth.CheckRecentTxs(ctx, client, int(depth), common.Address(batchSubmitterAddress))
	}
	require.NoErrorf(t, err, "failed when checking chain for recent batcher txs from %s", batchSubmitterAddress)
	require.True(t, found)
}
