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
	isExcluded := map[uint64]bool{
		291:          true, // incorrect scalar parameter
		888:          true, // no public endpoint
		957:          true, // incorrect scalar parameter
		997:          true, // no public endpoint
		58008:        true, // incorrect overhead parameter
		84532:        true, // incorrect overhead parameter
		11155421:     true, // no public endpoint
		11763071:     true, // no public endpoint
		11763072:     true, // no public endpoint
		129831238013: true, // no ground truth
	}

	checkResourceConfig := func(t *testing.T, chain *ChainConfig) {
		rpcEndpoint := chain.PublicRPC
		require.NotEmpty(t, rpcEndpoint, "no public endpoint for chain")

		client, err := ethclient.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		contractAddress := predeploys.GasPriceOracleAddr

		var desiredParams GasPriceOracleParams
		switch chain.Superchain {
		case "mainnet":
			desiredParams = OPMainnetGasPriceOracleParams
		case "goerli":
			desiredParams = OPGoerliGasPriceOracleParams
		case "sepolia":
			desiredParams = OPSepoliaGasPriceOracleParams
		case "goerli-dev-0":
			t.Fatalf("no ground truth for superchain %s", chain.Superchain)
		case "sepolia-dev-0":
			t.Fatalf("no ground truth for superchain %s", chain.Superchain)
		default:
			t.Fatalf("superchain not recognized: %s", chain.Superchain)
		}

		actualParams, err := getGasPriceOracleParamsWithRetries(context.Background(), contractAddress, client)
		require.NoErrorf(t, err, "RPC endpoint %s: %s", rpcEndpoint)

		require.Condition(t, func() bool { return (actualParams.Decimals.Cmp(desiredParams.Decimals) == 0) },
			"incorrect decimals parameter: got %d, wanted %d", actualParams.Decimals, desiredParams.Decimals)
		require.Condition(t, func() bool { return (actualParams.Overhead.Cmp(desiredParams.Overhead) == 0) },
			"incorrect overhead parameter: got %d, wanted %d", actualParams.Overhead, desiredParams.Overhead)
		require.Condition(t, func() bool { return (actualParams.Scalar.Cmp(desiredParams.Scalar) == 0) },
			"incorrect scalar parameter: got %d, wanted %d", actualParams.Scalar, desiredParams.Scalar)

		t.Logf("gas price oracle params are acceptable")

	}

	for chainID, chain := range OPChains {
		if !isExcluded[chainID] {
			t.Run(chain.Name+fmt.Sprintf(" (%d)", chainID), func(t *testing.T) { checkResourceConfig(t, chain) })
		}
	}
}

// getGasPriceOracleParamsWithRetries get the params stored in the contract at addr.
func getGasPriceOracleParamsWithRetries(ctx context.Context, addr common.Address, client *ethclient.Client) (GasPriceOracleParams, error) {
	maxAttempts := 3
	gasPriceOracle, err := bindings.NewGasPriceOracle(addr, client)
	if err != nil {
		return GasPriceOracleParams{}, fmt.Errorf("%s: %w", addr, err)
	}

	decimals, err := retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) { return gasPriceOracle.Decimals(&bind.CallOpts{Context: ctx}) })
	if err != nil {
		return GasPriceOracleParams{}, fmt.Errorf("%s.Decimals(): %w", addr, err)
	}

	overhead, err := retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) { return gasPriceOracle.Overhead(&bind.CallOpts{Context: ctx}) })
	if err != nil {
		return GasPriceOracleParams{}, fmt.Errorf("%s.Overhead(): %w", addr, err)
	}

	scalar, err := retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) { return gasPriceOracle.Scalar(&bind.CallOpts{Context: ctx}) })
	if err != nil {
		return GasPriceOracleParams{}, fmt.Errorf("%s.Scalar(): %w", addr, err)
	}

	return GasPriceOracleParams{
		Decimals: decimals, Overhead: overhead, Scalar: scalar,
	}, nil
}
