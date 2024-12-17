package validation

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/internal/bindings"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func testGasLimit(t *testing.T, chain *superchain.ChainConfig) {
	rpcEndpoint := superchain.Superchains[chain.Superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint)

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	err = CheckGasLimit(chain, client)
	require.NoError(t, err)
}

// Allows mock implementations to be passed in for testing
type getGasLimitWithRetriesFuncType func(context.Context, common.Address, *ethclient.Client) (uint64, error)

func CheckGasLimit(chain *superchain.ChainConfig, l1Client *ethclient.Client, override ...getGasLimitWithRetriesFuncType) error {

	f := getGasLimitWithRetries

	if len(override) > 0 {
		f = override[0]
	}

	contractAddress, err := superchain.Addresses[chain.ChainID].AddressFor("SystemConfigProxy")
	if err != nil {
		return err
	}

	desiredParam := standard.Config.Params[chain.Superchain].SystemConfig.GasLimit
	actualParam, err := f(context.Background(), common.Address(contractAddress), l1Client)
	if err != nil {
		return err
	}

	if !isIntWithinBounds(actualParam, desiredParam) {
		return fmt.Errorf("%d is not within bounds %d", actualParam, desiredParam)
	}
	return nil
}

func getGasLimitWithRetries(ctx context.Context, systemConfigAddr common.Address, client *ethclient.Client) (uint64, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	systemConfig, err := bindings.NewSystemConfig(systemConfigAddr, client)
	if err != nil {
		return 0, err
	}
	return Retry(systemConfig.GasLimit)(callOpts)
}
