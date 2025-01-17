package validation

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	offChainParam := chain.Genesis.L1.Number
	onChainParam, err := getStartBlockWithRetries(ctx, common.Address(systemConfigAddress), offChainParam, client)
	require.NoError(t, err)

	if offChainParam > onChainParam {
		// Ensure there aren't any skipped deposits in the block gap
		portalAddress := Addresses[chain.ChainID].OptimismPortalProxy
		require.NotZero(t, portalAddress)

		missedEvents, err := checkForDepositEvents(ctx, client, common.Address(portalAddress), onChainParam, offChainParam)
		require.NoError(t, err)
		require.Len(t, missedEvents, 0)
	}
}

func getStartBlockWithRetries(ctx context.Context, systemConfigAddr common.Address, genesisBlock uint64, client *ethclient.Client) (uint64, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	systemConfig, err := bindings.NewSystemConfig(systemConfigAddr, client)
	if err != nil {
		return 0, err
	}
	val, err := Retry(systemConfig.StartBlock)(callOpts)
	if err != nil {
		// Checking when the contract has been deployed, backward compatibility with SystemConfig without StartBlock()
		// Fetch the block when the proxy has been deployed via the AdminChanged event from 0x0 to the actual admin
		logs, err := client.FilterLogs(
			ctx,
			ethereum.FilterQuery{
				FromBlock: big.NewInt(int64(genesisBlock)), // we assume SystemConfig has been deployed closely to genesis (100 blocks around)
				ToBlock:   big.NewInt(int64(genesisBlock + 100)),
				Addresses: []common.Address{systemConfigAddr},
				Topics:    [][]common.Hash{{common.HexToHash("0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f")}}, // AdminChanged hash
			},
		)
		if err != nil {
			return 0, err
		}

		if len(logs) == 0 {
			return 0, fmt.Errorf("unable to find logs")
		} else if len(logs) > 1 {
			return 0, fmt.Errorf("multiple logs found")
		}

		log := logs[0]

		return log.BlockNumber, nil
	}

	return val.Uint64(), nil
}

// checkForDepositEvents looks in the blocks between startBlock and endBlock (inclusive) for any TransactionDeposited
// events emitted by the OptimismPortalProxy contract at portalAddress
func checkForDepositEvents(
	ctx context.Context,
	client *ethclient.Client,
	portalAddress common.Address,
	startBlock uint64,
	endBlock uint64,
) ([]types.Log, error) {
	// eventTopic for TransactionDeposited(address indexed from, address indexed to, uint256 indexed version, bytes opaqueData)
	eventTopic := common.HexToHash("0xb3813568d9991fc951961fcb4c784893574240a28925604d09fc577c55bb7c32")

	// Create a query
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(startBlock)),
		ToBlock:   big.NewInt(int64(endBlock)),
		Addresses: []common.Address{portalAddress},
		Topics:    [][]common.Hash{{eventTopic}},
	}

	missedEvents, err := client.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}

	return missedEvents, nil
}
