package validation

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"os/exec"
	"strings"
	"testing"

	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type EthClient interface {
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
}

func getAddressFromConfig(chainID uint64, contractName string) (superchain.Address, error) {
	if common.IsHexAddress(contractName) {
		return superchain.Address(common.HexToAddress(contractName)), nil
	}

	contractAddress, err := superchain.Addresses[chainID].AddressFor(contractName)

	return contractAddress, err
}

func getAddressFromChain(method string, contractAddress superchain.Address, client EthClient) (superchain.Address, error) {
	addr := (common.Address(contractAddress))
	callMsg := ethereum.CallMsg{
		To:   &addr,
		Data: crypto.Keccak256([]byte(method))[:4],
	}

	// Make the call
	callContract := func(msg ethereum.CallMsg) ([]byte, error) {
		return client.CallContract(context.Background(), msg, nil)
	}
	result, err := Retry(callContract)(callMsg)
	if err != nil {
		return superchain.Address{}, err
	}

	return superchain.Address(common.BytesToAddress(result)), nil
}

var checkResolutions = func(t *testing.T, r standard.Resolutions, chainID uint64, client *ethclient.Client) {
	for contract, methodToOutput := range r {
		contractAddress, err := getAddressFromConfig(chainID, contract)
		require.NoError(t, err)

		for method, output := range methodToOutput {
			want, err := getAddressFromConfig(chainID, output)
			require.NoError(t, err)

			got, err := getAddressFromChain(method, contractAddress, client)
			require.NoErrorf(t, err, "problem calling %s.%s (%s)", contract, method, contractAddress)

			// Use t.Errorf here for a concise output of failures, since failure info is sent to a slack channel
			if want != got {
				t.Errorf("%s.%s = %s, expected %s (%s)", contract, method, got, want, output)
			}
		}
	}
}

// isBigIntWithinBounds returns true if actual is within bounds, where the bounds are [lower bound, upper bound] and are inclusive.
var isBigIntWithinBounds = func(actual *big.Int, bounds [2]*big.Int) bool {
	if (bounds[1].Cmp(bounds[0])) < 0 {
		panic("bounds are in wrong order")
	}
	return (actual.Cmp(bounds[0]) >= 0 && actual.Cmp(bounds[1]) <= 0)
}

// isIntWithinBounds returns true if actual is within bounds, where the bounds are [lower bound, upper bound] and are inclusive.
func isIntWithinBounds[T uint32 | uint64](actual T, bounds [2]T) bool {
	if bounds[1] < bounds[0] {
		panic("bounds are in wrong order")
	}
	return (actual >= bounds[0] && actual <= bounds[1])
}

// assertInBounds fails the test (but not immediately) if the passed param is outside of the passed bounds.
func assertIntInBounds[T uint32 | uint64](t *testing.T, name string, got T, want [2]T) {
	assert.True(t,
		isIntWithinBounds(got, want),
		fmt.Sprintf("Incorrect %s, %d is not within bounds %d", name, got, want))
}

func CastCall(contractAddress superchain.Address, calldata string, args []string, rpcUrl string) ([]string, error) {
	cmdArgs := []string{"call", contractAddress.String(), calldata}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "-r", rpcUrl)

	cmd := exec.Command("cast", cmdArgs...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s, %w", &stderr, err)
	}

	results := strings.Fields(out.String())
	if results == nil {
		return nil, fmt.Errorf("cast call returned empty address")
	}

	return results, nil
}

const DefaultMaxRetries = 3

func Retry[S, T any](fn func(S) (T, error)) func(S) (T, error) {
	return func(s S) (T, error) {
		return retry.Do(context.Background(), DefaultMaxRetries, retry.Exponential(), func() (T, error) {
			return fn(s)
		})
	}
}
