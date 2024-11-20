package validation

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum-optimism/superchain-registry/superchain"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
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

	depth := chain.SequencerWindowSize
	require.NotZero(t, depth)

	blockNum, found, err := retry.Do2(context.Background(), DefaultMaxRetries, retry.Exponential(), func() (uint64, bool, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()
		// First attempt without needing an archive node
		//  - full nodes keep the last 128 blocks but we use 120 to give a buffer
		nonArchivalDepth := 120
		blockNum, found, err := eth.CheckRecentTxs(ctx, client, nonArchivalDepth, common.Address(batchSubmitterAddress))
		if !found {
			if err != nil {
				t.Log("failed to check recent batcher txs, will attempt retry")
			}
			blockNum, found, err = eth.CheckRecentTxs(ctx, client, int(depth), common.Address(batchSubmitterAddress))
		}

		return blockNum, found, err
	})
	require.NoErrorf(t, err, "failed when checking chain for recent batcher txs from %s", batchSubmitterAddress)
	require.True(t, found, "failed to find recent batcher tx")

	// Fetch the block by number
	block, err := client.BlockByNumber(context.Background(), new(big.Int).SetUint64(blockNum))
	require.NoError(t, err)

	// Iterate over the transactions in the block
	var chainConfig *params.ChainConfig
	if chain.Superchain == "mainnet" {
		chainConfig = params.MainnetChainConfig
	} else if chain.Superchain == "sepolia" || chain.Superchain == "sepolia-dev-0" {
		chainConfig = params.SepoliaChainConfig
	} else {
		require.Fail(t, "invalid l1 chain configured for l2 chain: %s", chain.Superchain)
	}
	for _, tx := range block.Transactions() {
		signer := types.MakeSigner(chainConfig, block.Number(), block.Time())
		sender, err := types.Sender(signer, tx)
		require.NoError(t, err)
		if sender == common.Address(batchSubmitterAddress) {
			require.Equal(t, *tx.To(), common.Address(batchInboxAddress))
			return
		}
	}
	require.Fail(t, "failed to find tx from batch submitter to batch inbox")
}
