package youplus

import (
	"context"
	youplustoolkitrpc "github.com/project-xpolaris/youplustoolkit/youplus/rpc"
	"time"
	"youfile/config"
)

var DefaultYouPlusRPCClient *youplustoolkitrpc.YouPlusRPCClient

func InitRPCClient() error {
	DefaultYouPlusRPCClient = youplustoolkitrpc.NewYouPlusRPCClient(config.Instance.YouPlusRPC)
	timeoutCtx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return DefaultYouPlusRPCClient.Connect(timeoutCtx)
}
