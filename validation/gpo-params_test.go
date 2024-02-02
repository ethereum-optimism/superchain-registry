package validation

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-service/retry"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestGasPriceOracleParams(t *testing.T) {
	isExcluded := map[uint64]bool{}

	checkResourceConfig := func(t *testing.T, chain *ChainConfig) {
		rpcEndpoint := chain.PublicRPC
		require.NotEmpty(t, rpcEndpoint, "no public endpoint for chain")

		client, err := ethclient.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		contractAddress := predeploys.GasPriceOracleAddr

		desiredParams := gasPriceOracleParams{
			decimals: big.NewInt(6),
			overhead: big.NewInt(188),
			scalar:   big.NewInt(684000),
		} // from OP Mainnet

		actualParams, err := getGasPriceOracleParamsWithRetries(context.Background(), contractAddress, client)
		require.NoErrorf(t, err, "RPC endpoint %s: %s", rpcEndpoint)

		require.Condition(t, func() bool { return (actualParams.decimals.Cmp(desiredParams.decimals) == 0) },
			"incorrect decimals parameter: got %d, wanted %d", actualParams.decimals, desiredParams.decimals)
		require.Condition(t, func() bool { return (actualParams.overhead.Cmp(desiredParams.overhead) == 0) },
			"incorrect overhead parameter: got %d, wanted %d", actualParams.overhead, desiredParams.overhead)
		require.Condition(t, func() bool { return (actualParams.scalar.Cmp(desiredParams.scalar) == 0) },
			"incorrect scalar parameter: got %d, wanted %d", actualParams.scalar, desiredParams.scalar)

		t.Logf("gas price oracle params are acceptable")

	}

	for chainID, chain := range OPChains {
		if !isExcluded[chainID] {
			t.Run(chain.Name+fmt.Sprintf(" (%d)", chainID), func(t *testing.T) { checkResourceConfig(t, chain) })
		}
	}
}

type gasPriceOracleParams struct {
	decimals *big.Int
	overhead *big.Int
	scalar   *big.Int
}

// getGasPriceOracleParamsWithRetries get the params stored in the contract at addr.
func getGasPriceOracleParamsWithRetries(ctx context.Context, addr common.Address, client *ethclient.Client) (gasPriceOracleParams, error) {
	maxAttempts := 10
	gasPriceOracle, err := bindings.NewGasPriceOracle(addr, client)
	if err != nil {
		return gasPriceOracleParams{}, fmt.Errorf("%s: %w", addr, err)
	}

	decimals, err := retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) { return gasPriceOracle.Decimals(&bind.CallOpts{Context: ctx}) })
	if err != nil {
		return gasPriceOracleParams{}, fmt.Errorf("%s.decimals(): %w", addr, err)
	}

	overhead, err := retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) { return gasPriceOracle.Overhead(&bind.CallOpts{Context: ctx}) })
	if err != nil {
		return gasPriceOracleParams{}, fmt.Errorf("%s.overhead(): %w", addr, err)
	}

	scalar, err := retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) { return gasPriceOracle.Scalar(&bind.CallOpts{Context: ctx}) })
	if err != nil {
		return gasPriceOracleParams{}, fmt.Errorf("%s.scalar(): %w", addr, err)
	}

	return gasPriceOracleParams{
		decimals, overhead, scalar,
	}, nil
}
