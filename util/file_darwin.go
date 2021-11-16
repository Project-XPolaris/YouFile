package util

import "os"

func ReadDisks() ([]string, error) {
	return nil, nil
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
	return directories
}
