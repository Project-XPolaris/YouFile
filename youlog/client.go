package youlog

import (
	"fmt"
	"github.com/project-xpolaris/youplustoolkit/youlog"
	"github.com/rs/xid"
	"youfile/config"
	"youfile/youplus"
)

var DefaultLogger *youlog.LogClient

func InitLogService() error {
	DefaultLogger = &youlog.LogClient{}
	instance := config.Instance.YouLog.Instance
	if len(instance) == 0 {
		instance = fmt.Sprintf("%s_%s", config.Instance.YouLog.Application, xid.New().String())
	}
	DefaultLogger.Init(config.Instance.YouLog.Addr, config.Instance.YouLog.Application, instance)
	if config.Instance.YouLog.Remote {
		err := DefaultLogger.Connect(youplus.GenerateRPCTimeoutContext())
		if err != nil {
			return err
		}
		DefaultLogger.StartDaemon(config.Instance.YouLog.Retry)
	}
	return nil
}
