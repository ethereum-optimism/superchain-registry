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
		291:          true, // mainnet/orderly                 (incorrect scalar parameter)
		888:          true, // goerli-dev-0/op-labs-chaosnet-0 (no public endpoint)
		957:          true, // mainnet/lyra                    (incorrect scalar parameter)
		997:          true, // goerli-dev-0/op-labs-devnet-0   (no public endpoint)
		58008:        true, // sepolia/pgn                     (incorrect overhead parameter)
		84532:        true, // sepolia/base                    (incorrect overhead parameter)
		11155421:     true, // sepolia-dev-0/oplabs-devnet-0   (no public endpoint)
		11763071:     true, // goerli-dev-0/base-devnet-0      (no public endpoint)
		11763072:     true, // sepolia-dev-0/base-devnet-0     (no public endpoint)
		129831238013: true, // goerli-dev-0/conduit-devnet-0   (no ground truth)
	}

	gasPriceOraclAddr := predeploys.GasPriceOracleAddr

	checkPreEcotoneResourceConfig := func(t *testing.T, chain *ChainConfig, client *ethclient.Client) {

		var desiredParams PreEcotoneGasPriceOracleParams
		switch chain.Superchain {
		case "mainnet":
			desiredParams = OPMainnetPreEcotoneGasPriceOracleParams
		case "goerli":
			desiredParams = OPGoerliPreEcotoneGasPriceOracleParams
		case "sepolia":
			desiredParams = OPSepoliaPreEcotoneGasPriceOracleParams
		case "goerli-dev-0":
			t.Fatalf("no ground truth for superchain %s", chain.Superchain)
		case "sepolia-dev-0":
			t.Fatalf("no ground truth for superchain %s", chain.Superchain)
		default:
			t.Fatalf("superchain not recognized: %s", chain.Superchain)
		}

		actualParams, err := getPreEcotoneGasPriceOracleParams(context.Background(), gasPriceOraclAddr, client)
		require.NoError(t, err)

		require.Equal(t, actualParams.Decimals.Cmp(desiredParams.Decimals), 0,
			"incorrect decimals parameter: got %d, wanted %d", actualParams.Decimals, desiredParams.Decimals)
		require.Equal(t, actualParams.Overhead.Cmp(desiredParams.Overhead), 0,
			"incorrect overhead parameter: got %d, wanted %d", actualParams.Overhead, desiredParams.Overhead)
		require.Equal(t, actualParams.Scalar.Cmp(desiredParams.Scalar), 0,
			"incorrect scalar parameter: got %d, wanted %d", actualParams.Scalar, desiredParams.Scalar)

		t.Logf("gas price oracle params are acceptable")

	}

	checkEcotoneResourceConfig := func(t *testing.T, chain *ChainConfig, client *ethclient.Client) {

		var desiredParams EcotoneGasPriceOracleParams
		switch chain.Superchain {
		case "goerli":
			desiredParams = OPGoerliEcotoneGasPriceOracleParams
		case "sepolia":
		case "mainnet":
		case "goerli-dev-0":
		case "sepolia-dev-0":
			t.Fatalf("no ground truth for superchain %s", chain.Superchain)
		default:
			t.Fatalf("superchain not recognized: %s", chain.Superchain)
		}

		actualParams, err := getEcotoneGasPriceOracleParams(context.Background(), gasPriceOraclAddr, client)
		require.NoError(t, err)

		require.Equal(t, actualParams.Decimals.Cmp(desiredParams.Decimals), 0,
			"incorrect decimals parameter: got %d, wanted %d", actualParams.Decimals, desiredParams.Decimals)
		require.Equal(t, actualParams.BlobBaseFeeScalar, desiredParams.BlobBaseFeeScalar,
			"incorrect blobBaseFeeScalar parameter: got %d, wanted %d", actualParams.BlobBaseFeeScalar, desiredParams.BlobBaseFeeScalar)
		require.Equal(t, actualParams.BaseFeeScalar, desiredParams.BaseFeeScalar,
			"incorrect baseFeeScalar: got %d, wanted %d", actualParams.BaseFeeScalar, desiredParams.BaseFeeScalar)

		t.Logf("gas price oracle params are acceptable")

	}

	checkResourceConfig := func(t *testing.T, chain *ChainConfig, client *ethclient.Client) {
		if Superchains[chain.Superchain].IsEcotone() {
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
