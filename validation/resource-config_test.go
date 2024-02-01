package validation

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/retry"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestResourceConfig(t *testing.T) {
	isExcluded := map[uint64]bool{
		888:          true,
		997:          true,
		11155421:     true,
		11763071:     true,
		129831238013: true,
	}

	checkResourceConfig := func(t *testing.T, chain *ChainConfig) {
		rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC

		require.NotEmpty(t, rpcEndpoint)

		client, err := ethclient.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		contractAddress, err := Addresses[chain.ChainID].AddressFor("SystemConfigProxy")
		require.NoError(t, err)

		uint128Max, ok := new(big.Int).SetString("ffffffffffffffffffffffffffffffff", 16)
		if !ok {
			panic("cannot construct uint128Max")
		}
		desiredResourceConfig := bindings.ResourceMeteringResourceConfig{
			MaxResourceLimit:            20000000,
			ElasticityMultiplier:        10,
			BaseFeeMaxChangeDenominator: 8,
			MinimumBaseFee:              1000000000,
			SystemTxMaxGas:              1000000,
			MaximumBaseFee:              uint128Max,
		} // from OP Mainnet

		actualResourceConfig, err := getResourceConfigWithRetries(context.Background(), common.Address(contractAddress), client)
		require.NoErrorf(t, err, "RPC endpoint %s: %s", rpcEndpoint)

		require.Equal(t, desiredResourceConfig, actualResourceConfig, "resource config unacceptable")

		t.Logf("resource metering acceptable")

	}

	for chainID, chain := range OPChains {
		if !isExcluded[chainID] {
			t.Run(chain.Name+fmt.Sprintf(" (%d)", chainID), func(t *testing.T) { checkResourceConfig(t, chain) })
		}
	}
}

// getResourceMeteringwill get the resoureConfig stored in the contract at systemConfigAddr.
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
