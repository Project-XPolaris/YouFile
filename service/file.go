package service

import (
	"github.com/spf13/afero"
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
func Copy(src string, dest string, notifier *CopyFileNotifier) error {
	srcStat, err := AppFs.Stat(src)
	if err != nil {
		return err
	}
	if srcStat.IsDir() {
		err = CopyDir(src, dest, notifier)
		if err != nil {
			return err
		}
	} else {
		err = CopyFile(src, dest, notifier)
		if err != nil {
			return err
		}
	}
	return nil
}
func Rename(oldName string, newName string) error {
	return AppFs.Rename(oldName, newName)
}
func DeleteFile(target string) error {
	return AppFs.RemoveAll(target)
}

func Chmod(target string, perm int) error {
	return AppFs.Chmod(target, os.FileMode(perm))
}

type CopyAnalyzeResult struct {
	FileCount int
	DirCount  int
	TotalSize int64
}

func analyzeSource(src string) (*CopyAnalyzeResult, error) {
	sourceStat, err := AppFs.Stat(src)
	if err != nil {
		return nil, err
	}
	if sourceStat.IsDir() {
		fileCounter := 0
		dirCount := 0
		var totalSize int64 = 0
		err := afero.Walk(AppFs, src, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				dirCount += 1
			} else {
				fileCounter += 1
			}
			totalSize += info.Size()
			return nil
		})
		return &CopyAnalyzeResult{FileCount: fileCounter, TotalSize: totalSize}, err
	} else {
		return &CopyAnalyzeResult{
			FileCount: 1,
			DirCount:  0,
			TotalSize: sourceStat.Size(),
		}, nil
	}
}

type DeleteNotifier struct {
	DeleteChan     chan string
	DeleteDoneChan chan string
	Info           *CopyAnalyzeResult
}

func Delete(src string, notifier *DeleteNotifier) error {
	srcStat, err := AppFs.Stat(src)
	if err != nil {
		return err
	}
	if srcStat.IsDir() {
		err = afero.Walk(AppFs, src, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				if notifier != nil {
					notifier.DeleteChan <- src
				}
				err := AppFs.Remove(path)
				//fmt.Println(path)
				if err != nil {
					return err
				}
				if notifier != nil {
					notifier.DeleteDoneChan <- src
				}
			}
			return nil
		})
		err = AppFs.RemoveAll(src)
	} else {
		if notifier != nil {
			notifier.DeleteChan <- src
		}
		err = AppFs.Remove(src)
		if notifier != nil {
			notifier.DeleteDoneChan <- src
		}
	}
	return err
}

func NewDirectory(dirPath string, perm int) error {
	err := AppFs.MkdirAll(dirPath, os.FileMode(perm))
	return err
}
