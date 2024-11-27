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
	"github.com/stretchr/testify/assert"
)

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
