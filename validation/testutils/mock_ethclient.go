package testutils

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/crypto"
)

type MockEthClient struct {
	// Response map keyed by method 4-byte signature
	Responses map[string][]byte
	Err       error
}

func (m *MockEthClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Responses[string(msg.Data)], nil
}

// Helper to get the 4-byte method ID
func MethodID(method string) []byte {
	return crypto.Keccak256([]byte(method))[:4]
}
