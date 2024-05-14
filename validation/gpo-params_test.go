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
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestGasPriceOracleParams(t *testing.T) {
	isExcluded := map[uint64]bool{
		11155421: true, // sepolia-dev-0/oplabs-devnet-0   (no public endpoint)
		11763072: true, // sepolia-dev-0/base-devnet-0     (no public endpoint)
	}

	gasPriceOraclAddr := predeploys.GasPriceOracleAddr

	checkPreEcotoneResourceConfig := func(t *testing.T, chain *ChainConfig, client *ethclient.Client) {
		desiredParams := standard.Config[chain.Superchain].GPOParams.PreEcotone

		actualParams, err := getPreEcotoneGasPriceOracleParams(context.Background(), gasPriceOraclAddr, client)
		require.NoError(t, err)

		assert.True(t, isBigIntWithinBounds(actualParams.Decimals, desiredParams.Decimals),
			"decimals parameter %d out of bounds %d", actualParams.Decimals, desiredParams.Decimals)
		assert.True(t, isBigIntWithinBounds(actualParams.Overhead, desiredParams.Overhead),
			"overhead parameter %d out of bounds %d", actualParams.Overhead, desiredParams.Overhead)
		assert.True(t, isBigIntWithinBounds(actualParams.Scalar, desiredParams.Scalar),
			"scalar parameter %d out of bounds %d", actualParams.Scalar, desiredParams.Scalar)
	}

	checkEcotoneResourceConfig := func(t *testing.T, chain *ChainConfig, client *ethclient.Client) {
		desiredParams := standard.Config[chain.Superchain].GPOParams.Ecotone

		actualParams, err := getEcotoneGasPriceOracleParams(context.Background(), gasPriceOraclAddr, client)
		require.NoError(t, err)

		assert.True(t, isBigIntWithinBounds(actualParams.Decimals, desiredParams.Decimals),
			"decimals parameter %d out of bounds %d", actualParams.Decimals, desiredParams.Decimals)
		assert.True(t, isIntWithinBounds(actualParams.BlobBaseFeeScalar, desiredParams.BlobBaseFeeScalar),
			"blobBaseFeeScalar %d out of bounds %d", actualParams.BlobBaseFeeScalar, desiredParams.BlobBaseFeeScalar)
		assert.True(t, isIntWithinBounds(actualParams.BaseFeeScalar, desiredParams.BaseFeeScalar),
			"baseFeeScalar parameter %d out of bounds %d", actualParams.BaseFeeScalar, desiredParams.BaseFeeScalar)
	}

	checkResourceConfig := func(t *testing.T, chain *ChainConfig, client *ethclient.Client) {
		if chain.IsEcotone() {
			checkEcotoneResourceConfig(t, chain, client)
		} else {
			checkPreEcotoneResourceConfig(t, chain, client)
		}
	}

	for chainID, chain := range OPChains {
		t.Run(perChainTestName(chain), func(t *testing.T) {
			if isExcluded[chainID] {
				t.Skip()
			}
			SkipCheckIfFrontierChain(t, *chain)
			rpcEndpoint := chain.PublicRPC
			require.NotEmpty(t, rpcEndpoint, "no public endpoint for chain")
			client, err := ethclient.Dial(rpcEndpoint)
			require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)
			checkResourceConfig(t, chain, client)
		})
	}
}

type PreEcotoneGasPriceOracleParams struct {
	Decimals *big.Int
	Overhead *big.Int
	Scalar   *big.Int
}

type EcotoneGasPriceOracleParams struct {
	Decimals          *big.Int
	BlobBaseFeeScalar uint32
	BaseFeeScalar     uint32
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
