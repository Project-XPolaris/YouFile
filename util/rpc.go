package util

import (
	"context"
	"time"
)

func GetRPCTimeout() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	return ctx
}
