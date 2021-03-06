package template

import (
	"github.com/d-tux/go-fstab"
	"path/filepath"
)

type MountTemplate struct {
	Spec      string            `json:"spec"`
	File      string            `json:"file"`
	VfsType   string            `json:"vfs_type"`
	MntOps    map[string]string `json:"mnt_ops"`
	Freq      int               `json:"freq"`
	PassNo    int               `json:"pass_no"`
	MountName string            `json:"mountName"`
}

func MountTemplateFromList(mounts fstab.Mounts) []MountTemplate {
	data := make([]MountTemplate, 0)
	for _, mount := range mounts {
		data = append(data, MountTemplate{
			Spec:      mount.Spec,
			File:      mount.File,
			VfsType:   mount.VfsType,
			MntOps:    mount.MntOps,
			Freq:      mount.Freq,
			PassNo:    mount.PassNo,
			MountName: filepath.Base(mount.File),
		})
	}
	return data
}
