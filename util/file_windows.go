package util

import (
	"os"
	"path/filepath"
	"syscall"
)

func ReadDisks() ([]string, error) {
	kernel32, _ := syscall.LoadLibrary("kernel32.dll")
	getLogicalDrivesHandle, _ := syscall.GetProcAddress(kernel32, "GetLogicalDrives")

	var drives []string

	if ret, _, callErr := syscall.Syscall(uintptr(getLogicalDrivesHandle), 0, 0, 0, 0); callErr != 0 {
		// handle error
	} else {
		drives = bitsToDrives(uint32(ret))
	}

	return drives, nil
}

func ReadStartDirectory() []*StartDirectory {
	directories := make([]*StartDirectory, 0)

	//user root
	homePath, err := os.UserHomeDir()
	if err == nil {
		directories = append(directories, &StartDirectory{
			Name: "Home",
			Path: homePath,
		})
	}
	directories = append(directories, &StartDirectory{
		Name: "Desktop",
		Path: filepath.Join(os.Getenv("USERPROFILE"), "Desktop"),
	})
	directories = append(directories, &StartDirectory{
		Name: "Downloads",
		Path: filepath.Join(os.Getenv("USERPROFILE"), "Downloads"),
	})
	directories = append(directories, &StartDirectory{
		Name: "Documents",
		Path: filepath.Join(os.Getenv("USERPROFILE"), "Documents"),
	})
	return directories
}
