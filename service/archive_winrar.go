package service

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type WinRARExtractOption struct {
	BaseExtractOption
}
type WinRARCompressOption struct {
	BaseCompressOption
}
type WinRAREngine struct {
	RarPath   string
	UnRARPath string
}

func (e *WinRAREngine) CanCompress() bool {
	return true
}

func (e *WinRAREngine) CanExtract() bool {
	return true
}

func NewWinRAREngine(rarPath string, unRARPath string) *WinRAREngine {
	return &WinRAREngine{RarPath: rarPath, UnRARPath: unRARPath}
}

func (e *WinRAREngine) Compress(target []string, output string, option CompressOption) error {
	args := []string{
		"a", "-y",
	}
	if len(option.GetPassword()) > 0 {
		args = append(args, fmt.Sprintf("-p%s", option.GetPassword()))
	}
	args = append(args, output)
	args = append(args, target...)
	cmd := exec.Command(e.RarPath, args...)
	cmd.Dir = filepath.Dir(output)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (e *WinRAREngine) Extract(target string, output string, option ExtractOption) error {
	args := []string{
		"x", "-y",
	}
	password := option.GetPassword()
	if len(password) > 0 {
		args = append(args, fmt.Sprintf("-p%s", password))
	}
	if !strings.HasSuffix(output, "\\") {
		output += "\\"
	}
	args = append(args, target, output)
	cmd := exec.Command(e.UnRARPath, args...)
	rawOutput := ""
	go func() {
		sout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Println(err)
		}
		scanner := bufio.NewScanner(sout)
		//scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
			rawOutput += fmt.Sprintf("%s\n", m)
		}
	}()
	err := cmd.Run()
	if err != nil {
		return err
	}
	fmt.Println(rawOutput)
	return nil
}
