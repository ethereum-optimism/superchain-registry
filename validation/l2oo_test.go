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

type l2OOParams struct {
	startingBlockNumber       *big.Int
	startingTimestamp         *big.Int
	submissionInterval        *big.Int
	l2BlockTime               *big.Int
	finalizationPeriodSeconds *big.Int
}

func TestL2OOParams(t *testing.T) {

	isExcluded := map[uint64]bool{
		291:          true,
		424:          true,
		888:          true,
		957:          true,
		997:          true,
		8453:         true,
		34443:        true,
		58008:        true,
		84531:        true,
		84532:        true,
		7777777:      true,
		11155421:     true, // sepolia-dev-0/oplabs-devnet-0
		11763071:     true,
		999999999:    true,
		129831238013: true,
	}

	checkEquality := func(a, b *big.Int) func() bool {
		return (func() bool { return (a.Cmp(b) == 0) })
	}

	incorrectMsg := func(name string, want, got *big.Int) string {
		return fmt.Sprintf("Incorrect %s, wanted %d got %d", name, want, got)

	}

	requireEqualParams := func(t *testing.T, desired, actual l2OOParams) {
		require.Condition(t,
			checkEquality(desired.startingBlockNumber, actual.startingBlockNumber),
			incorrectMsg("startingBlockNumber", desired.startingBlockNumber, actual.startingBlockNumber))
		require.Condition(t,
			checkEquality(desired.startingTimestamp, actual.startingTimestamp),
			incorrectMsg("startingTimestamp", desired.startingTimestamp, actual.startingTimestamp))
		require.Condition(t,
			checkEquality(desired.submissionInterval, actual.submissionInterval),
			incorrectMsg("submissionInterval", desired.submissionInterval, actual.submissionInterval))
		require.Condition(t,
			checkEquality(desired.l2BlockTime, actual.l2BlockTime),
			incorrectMsg("l2BlockTime", desired.l2BlockTime, actual.l2BlockTime))
		require.Condition(t,
			checkEquality(desired.finalizationPeriodSeconds, actual.finalizationPeriodSeconds),
			incorrectMsg("finalizationPeriodSeconds", desired.finalizationPeriodSeconds, actual.finalizationPeriodSeconds))
	}

	checkL2OOParams := func(t *testing.T, chain *ChainConfig) {
		rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC

		require.NotEmpty(t, rpcEndpoint)

		client, err := ethclient.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		contractAddress, err := Addresses[chain.ChainID].AddressFor("L2OutputOracleProxy")
		require.NoError(t, err)

		desiredParams := l2OOParams{
			startingBlockNumber:       big.NewInt(0),
			startingTimestamp:         big.NewInt(1690493568),
			submissionInterval:        big.NewInt(120),
			l2BlockTime:               big.NewInt(2),
			finalizationPeriodSeconds: big.NewInt(12),
		} // From OP Mainnet

		actualParams, err := getl2OOParamsWithRetries(context.Background(), common.Address(contractAddress), client)
		require.NoErrorf(t, err, "RPC endpoint %s", rpcEndpoint)

		requireEqualParams(t, desiredParams, actualParams)

		t.Logf("L2OutputOracle config params acceptable")

	}

	for chainID, chain := range OPChains {
		if !isExcluded[chainID] {
			t.Run(chain.Name+fmt.Sprintf(" (%d)", chainID), func(t *testing.T) { checkL2OOParams(t, chain) })
		}
	}
}

// getResourceMeteringwill gets each of the parameters from the L2OutputOracle at l2OOAddr,
// retrying up to 10 times with exponential backoff.
func getl2OOParamsWithRetries(ctx context.Context, l2OOAddr common.Address, client *ethclient.Client) (l2OOParams, error) {
	const maxAttempts = 10
	l2OO, err := bindings.NewL2OutputOracle(l2OOAddr, client)
	if err != nil {
		return l2OOParams{}, err
	}

	params := l2OOParams{}
	params.startingBlockNumber, err = retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) {
			return l2OO.StartingBlockNumber(&bind.CallOpts{Context: ctx})
		})
	if err != nil {
		return l2OOParams{}, fmt.Errorf("could not get startingBlockNumber: %w", err)
	}
	params.startingTimestamp, err = retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) {
			return l2OO.StartingTimestamp(&bind.CallOpts{Context: ctx})
		})
	if err != nil {
		return l2OOParams{}, fmt.Errorf("could not get startingTimestamp: %w", err)
	}
	params.submissionInterval, err = retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) {
			return l2OO.SubmissionInterval(&bind.CallOpts{Context: ctx})
		})
	if err != nil {
		return l2OOParams{}, fmt.Errorf("could not get submissionInterval: %w", err)
	}
	params.l2BlockTime, err = retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) {
			return l2OO.L2BlockTime(&bind.CallOpts{Context: ctx})
		})
	if err != nil {
		return l2OOParams{}, fmt.Errorf("could not get l2Blocktime: %w", err)
	}
	params.finalizationPeriodSeconds, err = retry.Do(ctx, maxAttempts, retry.Exponential(),
		func() (*big.Int, error) {
			return l2OO.FinalizationPeriodSeconds(&bind.CallOpts{Context: ctx})
		})
	if err != nil {
		return l2OOParams{}, fmt.Errorf("could not get finalizationPeriodSeconds: %w", err)
	}

	return params, nil
}
