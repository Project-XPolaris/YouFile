package service

import (
	"fmt"
	"os/user"
	"testing"
)

func TestFile(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		t.Error(err)
	}
	file, err := AppFs.Open(user.HomeDir)
	if err != nil {
		t.Error(err)
	}
	items, err := file.Readdir(0)
	if err != nil {
		t.Error(err)
	}
	for _, item := range items {
		fmt.Println(item.Name())
	}
}
