package validation

import (
	"context"

	"github.com/ethereum-optimism/optimism/op-service/retry"
)

func Retry[S, T any](fn func(S) (T, error)) func(S) (T, error) {
	const maxAttempts = 10
	return func(s S) (T, error) {
		return retry.Do(context.Background(), maxAttempts, retry.Exponential(), func() (T, error) {
			return fn(s)
		})
	}
}
