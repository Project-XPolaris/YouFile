package service

import (
	"fmt"
	"path/filepath"
	"runtime"
	"youfile/util"
)

type RootPath struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

func GetStartPath() ([]RootPath, error) {
	if runtime.GOOS != "windows" {
		return []RootPath{{Path: string(filepath.Separator), Name: "System Root"}}, nil
	}
	disks, err := util.ReadWindowsDisks()
	if err != nil {
		return nil, err
	}
	paths := make([]RootPath, 0)
	for _, disk := range disks {
		paths = append(paths, RootPath{
			Path: fmt.Sprintf("%s:\\", disk),
			Name: fmt.Sprintf("Disk %s", disk),
		})
	}
	return paths, err
}
