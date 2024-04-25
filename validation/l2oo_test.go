package validation

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	legacy "github.com/ethereum-optimism/superchain-registry/validation/internal/legacy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/retry"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type L2OOParams struct {
	SubmissionInterval        *big.Int `toml:"submission_interval"`         // Interval in blocks at which checkpoints must be submitted.
	L2BlockTime               *big.Int `toml:"l2_block_time"`               // The time per L2 block, in seconds.
	FinalizationPeriodSeconds *big.Int `toml:"finalization_period_seconds"` // The minimum time (in seconds) that must elapse before a withdrawal can be finalized.
}

func TestL2OOParams(t *testing.T) {
	isExcluded := map[uint64]bool{
		999999999: true, // sepolia/zora    Incorrect submissionInterval, wanted 120 got 180
		1740:      true, // sepolia/metal Incorrect submissionInterval
		919:       true, // sepolia/mode Incorrect submissionInterval
	}

	assertInBounds := func(t *testing.T, name string, got *big.Int, want [2]*big.Int) {
		assert.True(t,
			isBigIntWithinBounds(got, want),
			fmt.Sprintf("Incorrect %s, %d is not within bounds %d", name, got, want))
	}

	checkL2OOParams := func(t *testing.T, chain *ChainConfig) {
		rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC

		require.NotEmpty(t, rpcEndpoint)

		client, err := ethclient.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		contractAddress, err := Addresses[chain.ChainID].AddressFor("L2OutputOracleProxy")
		require.NoError(t, err)

		desiredParams := StandardConfig[chain.Superchain].L2OOParams

		version, err := getVersion(context.Background(), common.Address(contractAddress), client)
		require.NoError(t, err)

		var actualParams L2OOParams
		if version == "1.3.0" || version == "1.3.1" {
			actualParams, err = getl2OOParamsWithRetriesLegacy(context.Background(), common.Address(contractAddress), client)
		} else {
			actualParams, err = getl2OOParamsWithRetries(context.Background(), common.Address(contractAddress), client)
		}
		require.NoErrorf(t, err, "RPC endpoint %s", rpcEndpoint)

		assertInBounds(t, "submissionInterval", actualParams.SubmissionInterval, desiredParams.SubmissionInterval)
		assertInBounds(t, "l2BlockTime", actualParams.L2BlockTime, desiredParams.L2BlockTime)
		assertInBounds(t, "finalizationPeriodSeconds", actualParams.FinalizationPeriodSeconds, desiredParams.FinalizationPeriodSeconds)
	}

	for chainID, chain := range OPChains {
		t.Run(perChainTestName(chain), func(t *testing.T) {
			if isExcluded[chainID] {
				t.Skip()
			}
			SkipCheckIfFrontierChain(t, *chain)
			checkL2OOParams(t, chain)
		})
	}
}

// getResourceMeteringwill gets each of the parameters from the L2OutputOracle at l2OOAddr,
// retrying up to 10 times with exponential backoff.
func getl2OOParamsWithRetries(ctx context.Context, l2OOAddr common.Address, client *ethclient.Client) (L2OOParams, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	const maxAttempts = 3
	l2OO, err := bindings.NewL2OutputOracle(l2OOAddr, client)
	if err != nil {
		return L2OOParams{}, err
	}

	params := L2OOParams{}

	params.SubmissionInterval, err = retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) {
			return l2OO.SubmissionInterval(callOpts)
		})
	if err != nil {
		return L2OOParams{}, fmt.Errorf("could not get submissionInterval: %w", err)
	}
	params.L2BlockTime, err = retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) {
			return l2OO.L2BlockTime(callOpts)
		})
	if err != nil {
		return L2OOParams{}, fmt.Errorf("could not get l2Blocktime: %w", err)
	}
	params.FinalizationPeriodSeconds, err = retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) {
			return l2OO.FinalizationPeriodSeconds(callOpts)
		})
	if err != nil {
		return L2OOParams{}, fmt.Errorf("could not get finalizationPeriodSeconds: %w", err)
	}

	return params, nil
}

// getl2OOParamsWithRetriesLegacy gets each of the parameters from the L2OutputOracle at l2OOAddr,
// retrying up to 10 times with exponential backoff.
func getl2OOParamsWithRetriesLegacy(ctx context.Context, l2OOAddr common.Address, client *ethclient.Client) (L2OOParams, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	const maxAttempts = 3
	l2OO, err := legacy.NewL2OutputOracleCaller(l2OOAddr, client)
	if err != nil {
		return L2OOParams{}, err
	}

	params := L2OOParams{}

	params.SubmissionInterval, err = retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) {
			return l2OO.SUBMISSIONINTERVAL(callOpts)
		})
	if err != nil {
		return L2OOParams{}, fmt.Errorf("could not get submissionInterval: %w", err)
	}
	params.L2BlockTime, err = retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) {
			return l2OO.L2BLOCKTIME(callOpts)
		})
	if err != nil {
		return L2OOParams{}, fmt.Errorf("could not get l2Blocktime: %w", err)
	}
	params.FinalizationPeriodSeconds, err = retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) {
			return l2OO.FINALIZATIONPERIODSECONDS(callOpts)
		})
	if err != nil {
		return L2OOParams{}, fmt.Errorf("could not get finalizationPeriodSeconds: %w", err)
	}

	return params, nil
}
