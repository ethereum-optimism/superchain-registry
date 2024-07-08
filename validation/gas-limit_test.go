package validation

import (
	"context"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/internal/bindings"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func testGasLimit(t *testing.T, chain *ChainConfig) {
	rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC

	require.NotEmpty(t, rpcEndpoint)

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	contractAddress, err := Addresses[chain.ChainID].AddressFor("SystemConfigProxy")
	require.NoError(t, err)

	desiredParam := standard.Config.Params[chain.Superchain].SystemConfig.GasLimit
	actualParam, err := getGasLimitWithRetries(context.Background(), common.Address(contractAddress), client)
	require.NoError(t, err)

	assertIntInBounds(t, "gas_limit", actualParam, desiredParam)
}

func getGasLimitWithRetries(ctx context.Context, systemConfigAddr common.Address, client *ethclient.Client) (uint64, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	systemConfig, err := bindings.NewSystemConfig(systemConfigAddr, client)
	if err != nil {
		return 0, err
	}
	return Retry(systemConfig.GasLimit)(callOpts)
}
