package validation

import (
	"context"
	"time"
)

func getDefaultContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
