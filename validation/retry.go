package validation

import (
	"github.com/ethereum-optimism/optimism/op-service/retry"
)

func Retry[S, T any](fn func(S) (T, error)) func(S) (T, error) {
	const maxAttempts = 3
	ctx, _ := getDefaultContext()
	return func(s S) (T, error) {
		return retry.Do(ctx, maxAttempts, retry.Exponential(), func() (T, error) {
			return fn(s)
		})
	}
}
