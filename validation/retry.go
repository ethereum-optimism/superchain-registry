package validation

import (
	"context"

	"github.com/ethereum-optimism/optimism/op-service/retry"
)

const DefaultMaxRetries = 3

func Retry[S, T any](fn func(S) (T, error)) func(S) (T, error) {
	return func(s S) (T, error) {
		return retry.Do(context.Background(), DefaultMaxRetries, retry.Exponential(), func() (T, error) {
			return fn(s)
		})
	}
}
