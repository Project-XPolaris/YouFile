package service

import (
	"errors"
	"fmt"
	. "github.com/ahmetb/go-linq/v3"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"strings"
)

type ReadDirOption struct {
	OrderKey  string
	Order     string
	FileFirst bool
}

func ReadDir(readPath string, option ReadDirOption) ([]os.FileInfo, error) {
	file, err := AppFs.Open(readPath)
	if err != nil {
		return nil, err
	}
	items, err := file.Readdir(0)

	if len(option.OrderKey) == 0 {
		option.OrderKey = "name"
	}
	if len(option.Order) == 0 {
		option.Order = "asc"
	}

	if option.OrderKey == "name" {
		From(items).Sort(func(i, j interface{}) bool {
			if option.Order == "asc" {
				return strings.Compare(i.(os.FileInfo).Name(), j.(os.FileInfo).Name()) == -1
			} else {
				return strings.Compare(i.(os.FileInfo).Name(), j.(os.FileInfo).Name()) == 1
			}
		}).ToSlice(&items)
	}

	if option.OrderKey == "type" {
		filePart := make([]os.FileInfo, 0)
		dirPart := make([]os.FileInfo, 0)
		// pick up directory
		From(items).Where(func(i interface{}) bool {
			return i.(os.FileInfo).IsDir()
		}).ToSlice(&dirPart)

		// pick up file
		From(items).Where(func(i interface{}) bool {
			return !i.(os.FileInfo).IsDir()
		}).Sort(func(i, j interface{}) bool {
			if option.Order == "asc" {
				return strings.Compare(filepath.Ext(i.(os.FileInfo).Name()), filepath.Ext(j.(os.FileInfo).Name())) == -1
			} else {
				return strings.Compare(filepath.Ext(i.(os.FileInfo).Name()), filepath.Ext(j.(os.FileInfo).Name())) == 1
			}
		}).ToSlice(&filePart)

		result := make([]os.FileInfo, 0)
		if option.FileFirst {
			result = append(result, filePart...)
			result = append(result, dirPart...)
		} else {
			result = append(result, dirPart...)
			result = append(result, filePart...)

		}

		return result, nil
	}

	if option.OrderKey == "size" {
		filePart := make([]os.FileInfo, 0)
		dirPart := make([]os.FileInfo, 0)
		// pick up directory
		From(items).Where(func(i interface{}) bool {
			return i.(os.FileInfo).IsDir()
		}).ToSlice(&dirPart)

		// pick up file
		From(items).Where(func(i interface{}) bool {
			return !i.(os.FileInfo).IsDir()
		}).Sort(func(i, j interface{}) bool {
			if option.Order == "asc" {
				return i.(os.FileInfo).Size() < j.(os.FileInfo).Size()
			} else {
				return i.(os.FileInfo).Size() > j.(os.FileInfo).Size()
			}
		}).ToSlice(&filePart)

		result := make([]os.FileInfo, 0)
		if option.FileFirst {
			result = append(result, filePart...)
			result = append(result, dirPart...)
		} else {
			result = append(result, dirPart...)
			result = append(result, filePart...)

		}
		return result, nil
	}

	if option.OrderKey == "mod" {
		filePart := make([]os.FileInfo, 0)
		dirPart := make([]os.FileInfo, 0)
		// pick up directory
		From(items).Where(func(i interface{}) bool {
			return i.(os.FileInfo).IsDir()
		}).ToSlice(&dirPart)

		// pick up file
		From(items).Where(func(i interface{}) bool {
			return !i.(os.FileInfo).IsDir()
		}).Sort(func(i, j interface{}) bool {
			if option.Order == "asc" {
				return i.(os.FileInfo).ModTime().Before(j.(os.FileInfo).ModTime())
			} else {
				return !i.(os.FileInfo).ModTime().Before(j.(os.FileInfo).ModTime())
			}
		}).ToSlice(&filePart)

		result := make([]os.FileInfo, 0)
		if option.FileFirst {
			result = append(result, filePart...)
			result = append(result, dirPart...)
		} else {
			result = append(result, dirPart...)
			result = append(result, filePart...)

		}
		return result, nil
	}

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

var (
	DeleteInterrupt = errors.New("delete interrupt")
)

type DeleteNotifier struct {
	DeleteChan     chan string
	DeleteDoneChan chan string
	Info           *CopyAnalyzeResult
	StopFlag       bool
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
			if notifier != nil && notifier.StopFlag {
				fmt.Println("try to stop delete")
				return DeleteInterrupt
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
		if notifier != nil && notifier.StopFlag {
			return DeleteInterrupt
		}
	}
	return err
}

func NewDirectory(dirPath string, perm int) error {
	err := AppFs.MkdirAll(dirPath, os.FileMode(perm))
	return err
}
