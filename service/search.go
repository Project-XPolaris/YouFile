package service

import (
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"strings"
)

func SearchFile(src string, targetName string) ([]os.FileInfo, error) {
	result := make([]os.FileInfo, 0)
	err := afero.Walk(AppFs, src, func(path string, info os.FileInfo, err error) error {
		basename := filepath.Base(path)
		if strings.Contains(basename, targetName) {
			result = append(result, info)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
