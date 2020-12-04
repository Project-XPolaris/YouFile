package service

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

func CopyFile(source, dest string) error {
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

	// Create the destination file.
	dst, err := AppFs.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy the contents of the file.
	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	// Copy the mode if the user can't
	// open the file.
	info, err := AppFs.Stat(source)
	if err != nil {
		err = AppFs.Chmod(dest, info.Mode())
		if err != nil {
			return err
		}
	}
	<-time.After(10 * time.Second)

	return nil
}

func CopyDir(source, dest string) error {
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
			err = CopyDir(fsource, fdest)
			if err != nil {
				errs = append(errs, err)
			}
		} else {
			// Perform the file copy.
			err = CopyFile(fsource, fdest)
			if err != nil {
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
