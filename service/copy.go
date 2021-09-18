package service

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
	"youfile/util"
)

type CopyFileNotifier struct {
	CurrentFileChan   chan string
	CompleteDeltaChan chan int64
	FileCompleteChan  chan string
	StopChan          chan struct{}
	StopFlag          bool
}

func CopyFile(source, dest string, notifier *CopyFileNotifier, onDuplicate string) error {
	if notifier != nil {
		if notifier.StopFlag {
			return nil
		}
		notifier.CurrentFileChan <- source
	}
	// Open the source file.
	src, err := AppFs.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	// Makes the directory needed to create the dst
	// file.
	err = AppFs.MkdirAll(filepath.Dir(dest), 0666)
	if err != nil {
		return err
	}
	targetDest := dest
	if onDuplicate != "overwrite" {
		for {
			targetStat, _ := AppFs.Stat(targetDest)
			if targetStat != nil {
				if onDuplicate == "skip" {
					return nil
				}
			} else {
				break
			}
			targetDest = util.RenameDuplicateFilename(targetDest)
		}
	}

	// Create the destination file.
	dst, err := AppFs.OpenFile(targetDest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775)
	if err != nil {
		return err
	}
	defer dst.Close()
	srcStats, err := AppFs.Stat(source)
	if err != nil {
		return err
	}
	counterReader := util.NewCounterReader(src)
	var lastCompleteLength int64 = 0
	if notifier != nil {
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			for {
				select {
				case <-ticker.C:
					notifier.CompleteDeltaChan <- counterReader.N() - lastCompleteLength
					lastCompleteLength = counterReader.N()
					if counterReader.N() == srcStats.Size() {
						return
					}
				case <-notifier.StopChan:
					counterReader.StopChan <- struct{}{}
				}
			}
		}()
	}
	// Copy the contents of the file.
	_, err = io.Copy(dst, counterReader)

	if err != nil {
		return err
	}

	// Copy the mode if the user can't
	// open the file.
	info, err := AppFs.Stat(source)
	if err != nil {
		err = AppFs.Chmod(targetDest, info.Mode())
		if err != nil {
			return err
		}
	}

	if notifier != nil {
		notifier.FileCompleteChan <- source
	}
	return nil
}

func CopyDir(source, dest string, notifier *CopyFileNotifier, onDuplicate string) error {
	if notifier != nil && notifier.StopFlag {
		return nil
	}
	// Get properties of source.
	srcinfo, err := AppFs.Stat(source)
	if err != nil {
		return err
	}

	// Create the destination directory.
	err = AppFs.MkdirAll(dest, srcinfo.Mode())
	if err != nil {
		return err
	}

	dir, _ := AppFs.Open(source)
	obs, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	var errs []error

	for _, obj := range obs {
		fsource := source + "/" + obj.Name()
		fdest := dest + "/" + obj.Name()

		if obj.IsDir() {
			// Create sub-directories, recursively.
			err = CopyDir(fsource, fdest, notifier, onDuplicate)
			if err != nil {
				if err == util.CopyInterrupt {
					return err
				}
				errs = append(errs, err)
			}
		} else {
			// Perform the file copy.
			err = CopyFile(fsource, fdest, notifier, onDuplicate)
			if err != nil {
				if err == util.CopyInterrupt {
					return err
				}
				errs = append(errs, err)
			}
		}
	}

	var errString string
	for _, err := range errs {
		errString += err.Error() + "\n"
	}

	if errString != "" {
		return errors.New(errString)
	}

	return nil
}
