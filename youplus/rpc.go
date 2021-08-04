package youplus

import (
	youplustoolkitrpc "github.com/project-xpolaris/youplustoolkit/youplus/rpc"
	"youfile/config"
)

var DefaultYouPlusRPCClient *youplustoolkitrpc.YouPlusRPCClient

func InitRPCClient() error {
	DefaultYouPlusRPCClient = youplustoolkitrpc.NewYouPlusRPCClient(config.Instance.YouPlusRPC)
	return DefaultYouPlusRPCClient.Connect()
}
