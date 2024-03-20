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
		291:       true, // mainnet/orderly                 (incorrect scalar parameter)
		424:       true, // mainnet/pgn                     (blobBaseFeeScalar out of bounds)
		957:       true, // mainnet/lyra                    (incorrect scalar parameter)
		58008:     true, // sepolia/pgn                     (blobBaseFeeScalar out of bounds)
		84532:     true, // sepolia/base                    (blobBaseFeeScalar out of bounds)
		11155421:  true, // sepolia-dev-0/oplabs-devnet-0   (no public endpoint)
		11763072:  true, // sepolia-dev-0/base-devnet-0     (no public endpoint)
		999999999: true, // sepolia/zora                    (blobBaseFeeScalar out of bounds)
	}

	gasPriceOraclAddr := predeploys.GasPriceOracleAddr

	checkPreEcotoneResourceConfig := func(t *testing.T, chain *ChainConfig, client *ethclient.Client) {
		desiredParamsOuter, ok := GasPriceOracleParams[chain.Superchain]

		if !ok {
			t.Fatalf("superchain not recognized: %s", chain.Superchain)
		}
		desiredParams := desiredParamsOuter.PreEcotone

		actualParams, err := getPreEcotoneGasPriceOracleParams(context.Background(), gasPriceOraclAddr, client)
		require.NoError(t, err)

		require.True(t, isBigIntWithinBounds(actualParams.Decimals, desiredParams.Decimals.Bounds),
			"decimals parameter %d out of bounds %d", actualParams.Decimals, desiredParams.Decimals.Bounds)
		require.True(t, isBigIntWithinBounds(actualParams.Overhead, desiredParams.Overhead.Bounds),
			"overhead parameter %d out of bounds %d", actualParams.Overhead, desiredParams.Overhead.Bounds)
		require.True(t, isBigIntWithinBounds(actualParams.Scalar, desiredParams.Scalar.Bounds),
			"scalar parameter %d out of bounds %d", actualParams.Scalar, desiredParams.Scalar.Bounds)

		t.Logf("gas price oracle params are acceptable")
	}

	checkEcotoneResourceConfig := func(t *testing.T, chain *ChainConfig, client *ethclient.Client) {
		desiredParamsOuter, ok := GasPriceOracleParams[chain.Superchain]

		if !ok {
			t.Fatalf("superchain not recognized: %s", chain.Superchain)
		}
		desiredParams := desiredParamsOuter.Ecotone

		if desiredParams == nil {
			t.Fatal("no desiredParams.Ecotone set to compare Ecotone chain to")
		}

		actualParams, err := getEcotoneGasPriceOracleParams(context.Background(), gasPriceOraclAddr, client)
		require.NoError(t, err)

		require.True(t, isBigIntWithinBounds(actualParams.Decimals, desiredParams.Decimals.Bounds),
			"decimals parameter %d out of bounds %d", actualParams.Decimals, desiredParams.Decimals.Bounds)
		require.True(t, isWithinBounds(actualParams.BlobBaseFeeScalar, desiredParams.BlobBaseFeeScalar.Bounds),
			"blobBaseFeeScalar %d out of bounds %d", actualParams.BlobBaseFeeScalar, desiredParams.BlobBaseFeeScalar.Bounds)
		require.True(t, isWithinBounds(actualParams.BaseFeeScalar, desiredParams.BaseFeeScalar.Bounds),
			"baseFeeScalar parameter %d out of bounds %d", actualParams.BaseFeeScalar, desiredParams.BaseFeeScalar.Bounds)

		t.Logf("gas price oracle params are acceptable")
	}

	checkResourceConfig := func(t *testing.T, chain *ChainConfig, client *ethclient.Client) {
		if chain.IsEcotone() {
			checkEcotoneResourceConfig(t, chain, client)
		} else {
			checkPreEcotoneResourceConfig(t, chain, client)
		}
	}

	for chainID, chain := range OPChains {
		if !isExcluded[chainID] {
			t.Run(chain.Name+fmt.Sprintf(" (%d)", chainID), func(t *testing.T) {
				rpcEndpoint := chain.PublicRPC
				require.NotEmpty(t, rpcEndpoint, "no public endpoint for chain")
				client, err := ethclient.Dial(rpcEndpoint)
				require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)
				checkResourceConfig(t, chain, client)
			})
		}
	}
}

// getPreEcotoneGasPriceOracleParams gets the params by calling getters on the contract at addr. Will retry up to 3 times for each getter.
func getPreEcotoneGasPriceOracleParams(ctx context.Context, addr common.Address, client *ethclient.Client) (PreEcotoneGasPriceOracleParams, error) {
	maxAttempts := 3
	callOpts := &bind.CallOpts{Context: ctx}
	gasPriceOracle, err := bindings.NewGasPriceOracle(addr, client)
	if err != nil {
		return PreEcotoneGasPriceOracleParams{}, fmt.Errorf("%s: %w", addr, err)
	}

	decimals, err := retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) { return gasPriceOracle.Decimals(callOpts) })
	if err != nil {
		return PreEcotoneGasPriceOracleParams{}, fmt.Errorf("%s.Decimals(): %w", addr, err)
	}

	overhead, err := retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) { return gasPriceOracle.Overhead(callOpts) })
	if err != nil {
		return PreEcotoneGasPriceOracleParams{}, fmt.Errorf("%s.Overhead(): %w", addr, err)
	}

	scalar, err := retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) { return gasPriceOracle.Scalar(callOpts) })
	if err != nil {
		return PreEcotoneGasPriceOracleParams{}, fmt.Errorf("%s.Scalar(): %w", addr, err)
	}

	return PreEcotoneGasPriceOracleParams{
		Decimals: decimals, Overhead: overhead, Scalar: scalar,
	}, nil
}

// getEcotoneGasPriceOracleParams gets the params by calling getters on the contract at addr. Will retry up to 3 times for each getter.
func getEcotoneGasPriceOracleParams(ctx context.Context, addr common.Address, client *ethclient.Client) (EcotoneGasPriceOracleParams, error) {
	maxAttempts := 3
	callOpts := &bind.CallOpts{Context: ctx}
	gasPriceOracle, err := bindings.NewGasPriceOracle(addr, client)
	if err != nil {
		return EcotoneGasPriceOracleParams{}, fmt.Errorf("%s: %w", addr, err)
	}

	decimals, err := retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) { return gasPriceOracle.Decimals(callOpts) })
	if err != nil {
		return EcotoneGasPriceOracleParams{}, fmt.Errorf("%s.Decimals(): %w", addr, err)
	}

	blobBaseFeeScalar, err := retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (uint32, error) { return gasPriceOracle.BlobBaseFeeScalar(callOpts) })
	if err != nil {
		return EcotoneGasPriceOracleParams{}, fmt.Errorf("%s.BlobBaseFeeScalar(): %w", addr, err)
	}

	baseFeeScalar, err := retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (uint32, error) { return gasPriceOracle.BaseFeeScalar(callOpts) })
	if err != nil {
		return EcotoneGasPriceOracleParams{}, fmt.Errorf("%s.BaseFeeScalar(): %w", addr, err)
	}

	return EcotoneGasPriceOracleParams{
		Decimals:          decimals,
		BlobBaseFeeScalar: blobBaseFeeScalar,
		BaseFeeScalar:     baseFeeScalar,
	}, nil
}
