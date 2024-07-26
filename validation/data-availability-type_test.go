package validation

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/superchain-registry/superchain"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

	batchInboxAddress := chain.BatchInboxAddr
	require.NotZero(t, batchInboxAddress)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	depth := chain.SequencerWindowSize
	require.NotZero(t, depth)

	// First attempt without needing an archive node
	//  - full nodes keep the last 128 blocks but we use 120 to give a buffer
	nonArchivalDepth := 120
	blockNum, found, err := eth.CheckRecentTxs(ctx, client, nonArchivalDepth, common.Address(batchSubmitterAddress))
	if !found {
		blockNum, found, err = eth.CheckRecentTxs(ctx, client, int(depth), common.Address(batchSubmitterAddress))
	}
	require.NoErrorf(t, err, "failed when checking chain for recent batcher txs from %s", batchSubmitterAddress)
	require.True(t, found)

	// Fetch the block by number
	block, err := client.BlockByNumber(context.Background(), new(big.Int).SetUint64(blockNum))
	require.NoError(t, err)

	// Iterate over the transactions in the block
	for _, tx := range block.Transactions() {
		signer := types.NewCancunSigner(tx.ChainId())
		sender, err := types.Sender(signer, tx)
		require.NoError(t, err)
		if sender == common.Address(batchSubmitterAddress) {
			require.Equal(t, *tx.To(), common.Address(batchInboxAddress))
			return
		}
	}
	require.Fail(t, "failed to find tx from batch submitter to batch inbox")
}
