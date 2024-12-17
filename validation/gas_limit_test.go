package validation

import (
	"context"
	"errors"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestCheckGasLimit(t *testing.T) {
	mockClient := &ethclient.Client{}
	mockChain := &superchain.ChainConfig{
		ChainID:    10,
		Superchain: "mainnet",
	}

	// Restore the original function after the test
	originalFunc := getGasLimitWithRetriesFunc
	defer func() { getGasLimitWithRetriesFunc = originalFunc }()

	t.Run("Success", func(t *testing.T) {
		getGasLimitWithRetriesFunc = func(ctx context.Context, addr common.Address, client *ethclient.Client) (uint64, error) {
			return 30000000, nil // Gas limit within bounds
		}

		err := CheckGasLimit(mockChain, mockClient)
		require.NoError(t, err)
	})

	t.Run("ErrorGetGasLimit", func(t *testing.T) {
		getGasLimitWithRetriesFunc = func(ctx context.Context, addr common.Address, client *ethclient.Client) (uint64, error) {
			return 0, errors.New("failed to fetch gas limit")
		}

		err := CheckGasLimit(mockChain, mockClient)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to fetch gas limit")
	})

	t.Run("ErrorOutOfBounds", func(t *testing.T) {
		// Override the function to return a gas limit outside the desired bounds
		getGasLimitWithRetriesFunc = func(ctx context.Context, addr common.Address, client *ethclient.Client) (uint64, error) {
			return 150, nil // Out of bounds
		}

		err := CheckGasLimit(mockChain, mockClient)
		require.Error(t, err)
		require.Contains(t, err.Error(), "is not within bounds")
	})
}
