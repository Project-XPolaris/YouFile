package service

import "github.com/mholt/archiver/v3"

type DefaultArchiveEngine struct {
}
type DefaultArchiveExtractOption struct {
	BaseExtractOption
}

type DefaultArchiveCompressOption struct {
	BaseCompressOption
}

func (e *DefaultArchiveEngine) Compress(target []string, output string, option CompressOption) error {
	err := archiver.Archive(target, output)
	if err != nil {
		return err
	}
	return nil
}

func (e *DefaultArchiveEngine) Extract(target string, output string, option ExtractOption) error {
	err := archiver.Unarchive(target, output)
	return err
}
