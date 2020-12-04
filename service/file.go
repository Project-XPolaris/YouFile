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
func Copy(src string, dest string) error {
	srcStat, err := AppFs.Stat(src)
	if err != nil {
		return err
	}
	if srcStat.IsDir() {
		err = CopyDir(src, dest)
		if err != nil {
			return err
		}
	} else {
		err = CopyFile(src, dest)
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
