package service

import (
	"fmt"
	"os/exec"
)

type MountCIFSOption struct {
	MountPath  string `json:"mount_path"`
	RemotePath string `json:"remote_path"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

func MountCIFS(option MountCIFSOption) error {
	parts := make([]string, 0)
	parts = append(parts, "-t", "cifs")
	if len(option.Username) > 0 && len(option.Password) > 0 {
		parts = append(parts, "-o", fmt.Sprintf("username=%s,password=%s,dir_mode=0777,file_mode=0777", option.Username, option.Password))
	}
	parts = append(parts, option.RemotePath, option.MountPath)
	out, err := exec.Command("mount", parts...).Output()
	if err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("mount %s : %s", option.RemotePath, option.MountPath))
	output := string(out[:])
	fmt.Println(output)
	return nil
}

func UmountFS(dir string, extra ...string) error {
	params := make([]string, 0)
	params = append(params, dir)
	params = append(params, extra...)
	out, err := exec.Command("umount", params...).Output()
	if err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("umount %s", dir))
	output := string(out[:])
	fmt.Println(output)
	return nil
}
