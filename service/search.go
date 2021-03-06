package service

import (
	"errors"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"strings"
)

type SearchFileNotifier struct {
	HitChan  chan TargetFile
	StopFlag bool
}
type TargetFile struct {
	Path string
	Info os.FileInfo
}

var (
	StopWithLimit     = errors.New("at limit")
	StopWithInterrupt = errors.New("received interrupt")
)

func SearchFile(src string, targetName string, notifier *SearchFileNotifier, limit int) ([]TargetFile, error) {
	result := make([]TargetFile, 0)
	err := afero.Walk(AppFs, src, func(path string, info os.FileInfo, err error) error {
		basename := filepath.Base(path)
		if strings.Contains(basename, targetName) {
			if notifier != nil {
				notifier.HitChan <- TargetFile{
					Path: path,
					Info: info,
				}
			}
			result = append(result, TargetFile{
				Path: path,
				Info: info,
			})
			if limit != 0 {
				if len(result) == limit {
					return StopWithLimit
				}
			}
		}
		if notifier.StopFlag {
			//fmt.Println(" stop with interrupt")
			return StopWithInterrupt
		}
		return nil
	})
	if err != nil && err != StopWithLimit && err != StopWithInterrupt {
		return nil, err
	}
	return result, nil
}
