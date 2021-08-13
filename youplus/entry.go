package youplus

import (
	entry "github.com/project-xpolaris/youplustoolkit/youplus/entity"
	"youfile/config"
)

var DefaultEntry *entry.EntityClient

type AppExport struct {
	Addrs []string `json:"addrs"`
}

func InitEntity() {
	DefaultEntry = entry.NewEntityClient(config.Instance.Entity.Name, config.Instance.Entity.Version, &entry.EntityExport{}, DefaultYouPlusRPCClient)
	DefaultEntry.HeartbeatRate = 3000
}
