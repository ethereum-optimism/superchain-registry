package validation

import (
	"context"
	"fmt"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/internal/bindings"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func testStartBlock(t *testing.T, chain *ChainConfig) {
	rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint)

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	contractAddress := Addresses[chain.ChainID].SystemConfigProxy
	require.NotZero(t, contractAddress)

	offChainParam := chain.Genesis.L1.Number
	onChainParam, err := getStartBlockWithRetries(context.Background(), common.Address(contractAddress), client)
	require.NoError(t, err)

	isValid := offChainParam <= onChainParam
	// offChainParam determines when new nodes will start when trying to sync. If the offChain value is
	// greater than the onChain value, there is a risk new nodes will miss deposit txs
	require.True(t, isValid, fmt.Sprintf("off-chain = %d, on-chain = %d", offChainParam, onChainParam))
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
