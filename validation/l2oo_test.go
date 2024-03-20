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

func TestL2OOParams(t *testing.T) {
	isExcluded := map[uint64]bool{
		10:        true, // mainnet/op                      (old version of L2OutputOracle, no submissionInterval method)
		291:       true, // mainnet/orderly                 (old version of L2OutputOracle, no submissionInterval method)
		424:       true, // mainnet/pgn                     (old version of L2OutputOracle, no submissionInterval method)
		957:       true, // mainnet/lyra                    (old version of L2OutputOracle, no submissionInterval method)
		8453:      true, // mainnet/base                    (old version of L2OutputOracle, no submissionInterval method)
		34443:     true, // mainnet/mode                    (old version of L2OutputOracle, no submissionInterval method)
		58008:     true, // sepolia/pgn                     (old version of L2OutputOracle, no submissionInterval method)
		7777777:   true, // mainnet/zora                    (old version of L2OutputOracle, no submissionInterval method)
		11155421:  true, // sepolia-dev-0/oplabs-devnet-0   (old version of L2OutputOracle, no submissionInterval method)
		999999999: true, // sepolia/zora                    (old version of L2OutputOracle, no submissionInterval method)
	}

	checkEquality := func(a, b *big.Int) func() bool {
		return (func() bool { return (a.Cmp(b) == 0) })
	}

	incorrectMsg := func(name string, want, got *big.Int) string {
		return fmt.Sprintf("Incorrect %s, wanted %d got %d", name, want, got)
	}

	requireEqualParams := func(t *testing.T, desired, actual L2OOParams) {
		require.Condition(t,
			checkEquality(desired.SubmissionInterval, actual.SubmissionInterval),
			incorrectMsg("submissionInterval", desired.SubmissionInterval, actual.SubmissionInterval))
		require.Condition(t,
			checkEquality(desired.L2BlockTime, actual.L2BlockTime),
			incorrectMsg("l2BlockTime", desired.L2BlockTime, actual.L2BlockTime))
		require.Condition(t,
			checkEquality(desired.FinalizationPeriodSeconds, actual.FinalizationPeriodSeconds),
			incorrectMsg("finalizationPeriodSeconds", desired.FinalizationPeriodSeconds, actual.FinalizationPeriodSeconds))
	}

	checkL2OOParams := func(t *testing.T, chain *ChainConfig) {
		rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC

		require.NotEmpty(t, rpcEndpoint)

		client, err := ethclient.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		contractAddress, err := Addresses[chain.ChainID].AddressFor("L2OutputOracleProxy")
		require.NoError(t, err)

		var desiredParams L2OOParams
		switch chain.Superchain {
		case "mainnet":
			desiredParams = OPMainnetL2OOParams
		case "sepolia":
			desiredParams = OPSepoliaL2OOParams
		case "sepolia-dev-0":
			desiredParams = OPSepoliaDev0L2OOParams
		default:
			t.Fatalf("superchain not recognized: %s", chain.Superchain)
		}

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
