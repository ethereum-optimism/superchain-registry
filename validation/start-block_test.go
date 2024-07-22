package validation

import (
	"context"
	"math/big"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/internal/bindings"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func testStartBlock(t *testing.T, chain *ChainConfig) {
	rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint)

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	systemConfigAddress := Addresses[chain.ChainID].SystemConfigProxy
	require.NotZero(t, systemConfigAddress)

	offChainParam := chain.Genesis.L1.Number
	onChainParam, err := getStartBlockWithRetries(context.Background(), common.Address(systemConfigAddress), client)
	require.NoError(t, err)

	if offChainParam > onChainParam {
		// Ensure there aren't any skipped deposits in the block gap
		portalAddress := Addresses[chain.ChainID].OptimismPortalProxy
		require.NotZero(t, portalAddress)

		missedEvents, err := checkForDepositEvents(client, common.Address(portalAddress), onChainParam, offChainParam)
		require.NoError(t, err)
		require.Zero(t, len(missedEvents))
	}
}

func getStartBlockWithRetries(ctx context.Context, systemConfigAddr common.Address, client *ethclient.Client) (uint64, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	systemConfig, err := bindings.NewSystemConfig(systemConfigAddr, client)
	if err != nil {
		return 0, err
	}
	val, err := Retry(systemConfig.StartBlock)(callOpts)
	if err != nil {
		return 0, err
	}

	return val.Uint64(), nil
}

func checkForDepositEvents(client *ethclient.Client, portalAddress common.Address, startBlock uint64, endBlock uint64) ([]types.Log, error) {
	// eventTopic for TransactionDeposited(address indexed from, address indexed to, uint256 indexed version, bytes opaqueData)
	// - copied from validation/internal/bindings/optimism-portal.go
	eventTopic := common.HexToHash("0xb3813568d9991fc951961fcb4c784893574240a28925604d09fc577c55bb7c32")

	// Create a query
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(startBlock)),
		ToBlock:   big.NewInt(int64(endBlock)),
		Addresses: []common.Address{portalAddress},
		Topics:    [][]common.Hash{{eventTopic}},
	}

	// Fetch logs for the current chunk
	missedEvents, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		return []types.Log{}, err
	}

	return missedEvents, nil
}
