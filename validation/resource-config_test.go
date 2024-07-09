package validation

import (
	"context"
	"fmt"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/internal/bindings"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func testResourceConfig(t *testing.T, chain *ChainConfig) {
	rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC

	require.NotEmpty(t, rpcEndpoint)

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	contractAddress, err := Addresses[chain.ChainID].AddressFor("SystemConfigProxy")
	require.NoError(t, err)

	actualResourceConfig, err := getResourceConfig(context.Background(), common.Address(contractAddress), client)
	require.NoErrorf(t, err, "RPC endpoint %s: %s", rpcEndpoint)

	desiredParams := standard.Config.Params[chain.Superchain].ResourceConfig

	require.Equal(t, bindings.ResourceMeteringResourceConfig(desiredParams), actualResourceConfig, "resource config unacceptable")
}

// getResourceConfig will get the resoureConfig stored in the contract at systemConfigAddr.
func getResourceConfig(ctx context.Context, systemConfigAddr common.Address, client *ethclient.Client) (bindings.ResourceMeteringResourceConfig, error) {
	systemConfig, err := bindings.NewSystemConfig(systemConfigAddr, client)
	if err != nil {
		return bindings.ResourceMeteringResourceConfig{}, fmt.Errorf("%s: %w", systemConfigAddr, err)
	}

	resourceConfig, err := Retry(systemConfig.ResourceConfig)(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return bindings.ResourceMeteringResourceConfig{}, fmt.Errorf("%s: %w", systemConfigAddr, err)
	}

	return resourceConfig, nil
}
