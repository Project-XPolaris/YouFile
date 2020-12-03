package service

import (
	"os"
)

func ReadDir(readPath string) ([]os.FileInfo, error) {
	file, err := AppFs.Open(readPath)
	if err != nil {
		return nil, err
	}
	items, err := file.Readdir(0)
	return items, nil
}
