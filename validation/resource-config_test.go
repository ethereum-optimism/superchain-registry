package validation

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/retry"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestResourceConfig(t *testing.T) {
	checkResourceConfig := func(t *testing.T, chain *ChainConfig) {
		rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC

		require.NotEmpty(t, rpcEndpoint)

		client, err := ethclient.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		contractAddress, err := Addresses[chain.ChainID].AddressFor("SystemConfigProxy")
		require.NoError(t, err)

		actualResourceConfig, err := getResourceConfigWithRetries(context.Background(), common.Address(contractAddress), client)
		require.NoErrorf(t, err, "RPC endpoint %s: %s", rpcEndpoint)

		desiredParams := standard.Config[chain.Superchain].ResourceConfig

		require.Equal(t, bindings.ResourceMeteringResourceConfig(desiredParams), actualResourceConfig, "resource config unacceptable")
	}

	for _, chain := range OPChains {
		t.Run(perChainTestName(chain), func(t *testing.T) {
			SkipCheckIfFrontierChain(t, *chain)
			checkResourceConfig(t, chain)
		})
	}
}

// getResourceConfig will get the resoureConfig stored in the contract at systemConfigAddr.
func getResourceConfig(ctx context.Context, systemConfigAddr common.Address, client *ethclient.Client) (bindings.ResourceMeteringResourceConfig, error) {
	systemConfig, err := bindings.NewSystemConfig(systemConfigAddr, client)
	if err != nil {
		return bindings.ResourceMeteringResourceConfig{}, fmt.Errorf("%s: %w", systemConfigAddr, err)
	}

	resourceConfig, err := systemConfig.ResourceConfig(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return bindings.ResourceMeteringResourceConfig{}, fmt.Errorf("%s: %w", systemConfigAddr, err)
	}

	return resourceConfig, nil
}

// getResourceConfigWithRetries is a wrapper for getResourceMetering
// which retries up to 10 times with exponential backoff.
func getResourceConfigWithRetries(ctx context.Context, addr common.Address, client *ethclient.Client) (bindings.ResourceMeteringResourceConfig, error) {
	const maxAttempts = 10
	return retry.Do(ctx, maxAttempts, retry.Exponential(), func() (bindings.ResourceMeteringResourceConfig, error) {
		return getResourceConfig(ctx, addr, client)
	})
}
