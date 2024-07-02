package validation

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	legacy "github.com/ethereum-optimism/superchain-registry/validation/internal/legacy"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type L2OOParams struct {
	SubmissionInterval        *big.Int
	L2BlockTime               *big.Int
	FinalizationPeriodSeconds *big.Int
}

func testL2OOParams(t *testing.T, chain *ChainConfig) {
	isExcluded := map[uint64]bool{
		10:        true, // OP mainnet - Upgraded to fault proofs
		999999999: true, // sepolia/zora  Incorrect finalizationPeriodSeconds, 604800 is not within bounds [12 12]
		1740:      true, // sepolia/metal Incorrect finalizationPeriodSeconds, 604800 is not within bounds [12 12]
		919:       true, // sepolia/mode  Incorrect finalizationPeriodSeconds, 180 is not within bounds [12 12]
		11155420:  true, // sepolia/op No L2OO because this chain uses Fault Proofs https://github.com/ethereum-optimism/superchain-registry/issues/219
		11155421:  true, // oplabs-sepolia-devnet-0 No L2OO because this chain uses Fault Proofs https://github.com/ethereum-optimism/superchain-registry/issues/219
	}
	if isExcluded[chain.ChainID] {
		t.Skip()
	}

	rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC

	require.NotEmpty(t, rpcEndpoint)

	client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

	contractAddress, err := Addresses[chain.ChainID].AddressFor("L2OutputOracleProxy")
	require.NoError(t, err)

	desiredParams := standard.Config.Params[chain.Superchain].L2OOParams

	version, err := getVersion(context.Background(), common.Address(contractAddress), client)
	require.NoError(t, err)

	var actualParams L2OOParams
	if version == "1.3.0" || version == "1.3.1" {
		actualParams, err = getl2OOParamsWithRetriesLegacy(context.Background(), common.Address(contractAddress), client)
	} else {
		actualParams, err = getl2OOParamsWithRetries(context.Background(), common.Address(contractAddress), client)
	}
	require.NoErrorf(t, err, "RPC endpoint %s", rpcEndpoint)

	assertBigIntInBounds(t, "submissionInterval", actualParams.SubmissionInterval, desiredParams.SubmissionInterval)
	assertBigIntInBounds(t, "l2BlockTime", actualParams.L2BlockTime, desiredParams.L2BlockTime)
	assertBigIntInBounds(t, "challengePeriodSeconds", actualParams.FinalizationPeriodSeconds, desiredParams.ChallengePeriodSeconds)
}

// getl2OOParamsWithRetries gets each of the parameters from the L2OutputOracle at l2OOAddr,
// retrying up to 10 times with exponential backoff.
func getl2OOParamsWithRetries(ctx context.Context, l2OOAddr common.Address, client *ethclient.Client) (L2OOParams, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	l2OO, err := bindings.NewL2OutputOracle(l2OOAddr, client)
	if err != nil {
		return L2OOParams{}, err
	}

	params := L2OOParams{}

	params.SubmissionInterval, err = Retry(l2OO.SubmissionInterval)(callOpts)
	if err != nil {
		return L2OOParams{}, fmt.Errorf("could not get submissionInterval: %w", err)
	}

	params.L2BlockTime, err = Retry(l2OO.L2BlockTime)(callOpts)
	if err != nil {
		return L2OOParams{}, fmt.Errorf("could not get l2Blocktime: %w", err)
	}

	params.FinalizationPeriodSeconds, err = Retry(l2OO.FinalizationPeriodSeconds)(callOpts)
	if err != nil {
		return L2OOParams{}, fmt.Errorf("could not get finalizationPeriodSeconds: %w", err)
	}

	return params, nil
}

// getl2OOParamsWithRetriesLegacy gets each of the parameters from the L2OutputOracle at l2OOAddr,
func getl2OOParamsWithRetriesLegacy(ctx context.Context, l2OOAddr common.Address, client *ethclient.Client) (L2OOParams, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	l2OO, err := legacy.NewL2OutputOracleCaller(l2OOAddr, client)
	if err != nil {
		return L2OOParams{}, err
	}

	params := L2OOParams{}

	params.SubmissionInterval, err = Retry(l2OO.SUBMISSIONINTERVAL)(callOpts)
	if err != nil {
		return L2OOParams{}, fmt.Errorf("could not get submissionInterval: %w", err)
	}

	params.L2BlockTime, err = Retry(l2OO.L2BLOCKTIME)(callOpts)
	if err != nil {
		return L2OOParams{}, fmt.Errorf("could not get l2Blocktime: %w", err)
	}

	params.FinalizationPeriodSeconds, err = Retry(l2OO.FINALIZATIONPERIODSECONDS)(callOpts)
	if err != nil {
		return L2OOParams{}, fmt.Errorf("could not get finalizationPeriodSeconds: %w", err)
	}

	return params, nil
}
