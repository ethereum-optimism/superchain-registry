package testutils

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/stretchr/testify/mock"
)

type MockEthClient struct {
	mock.Mock
}

func (m *MockEthClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, msg, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}
