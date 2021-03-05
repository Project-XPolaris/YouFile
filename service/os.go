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
	Type string `json:"type"`
}

func GetStartPath() ([]RootPath, error) {
	if runtime.GOOS != "windows" {
		return []RootPath{{Path: string(filepath.Separator), Name: "System Root"}}, nil
	}
	disks, err := util.ReadDisks()
	if err != nil {
		return nil, err
	}
	paths := make([]RootPath, 0)
	for _, disk := range disks {
		paths = append(paths, RootPath{
			Path: fmt.Sprintf("%s:\\", disk),
			Name: fmt.Sprintf("Disk (%s)", disk),
			Type: "Parted",
		})
	}
	for _, directory := range util.ReadStartDirectory() {
		paths = append(paths, RootPath{
			Path: directory.Path,
			Name: directory.Name,
			Type: "Directory",
		})
	}
	return paths, err
}
