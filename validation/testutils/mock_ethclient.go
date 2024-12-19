package testutils

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/crypto"
)

// MockEthClient is a mock that implements EthCaller for testing.
type MockEthClient struct {
	// Setup response maps keyed by method 4-byte signature + address, or
	// just by method signature if that's simpler for your tests.
	Responses map[string][]byte
	Err       error
}

func (m *MockEthClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Responses[string(msg.Data)], nil
}

// Helper to get the 4-byte method ID (like what getBytes does internally)
func MethodID(method string) []byte {
	return crypto.Keccak256([]byte(method))[:4]
}
