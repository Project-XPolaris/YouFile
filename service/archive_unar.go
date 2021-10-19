package service

import "os/exec"

type UnarArchiveEngine struct {
}

func NewUnarArchiveEngine() *UnarArchiveEngine {
	return &UnarArchiveEngine{}
}

func (e *UnarArchiveEngine) CanCompress() bool {
	return false
}

func (e *UnarArchiveEngine) CanExtract() bool {
	return true
}

func (e *UnarArchiveEngine) Compress(target []string, output string, option CompressOption) error {
	return nil
}

func (e *UnarArchiveEngine) Extract(target string, output string, option ExtractOption) error {
	args := []string{
		"-o",
		output,
		"-q",
	}
	if len(option.GetPassword()) > 0 {
		args = append(args, "-p", option.GetPassword())
	}
	args = append(args, target)
	cmd := exec.Command("unar", args...)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

type UnarArchiveExtractOption struct {
	BaseExtractOption
}

type UnarArchiveCompressOption struct {
	BaseCompressOption
}
