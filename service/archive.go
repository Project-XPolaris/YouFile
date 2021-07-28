package service

import "youfile/config"

type ExtractOption interface {
	GetPassword() string
}
type CompressOption interface {
	GetPassword() string
}
type ArchiveEngine interface {
	Compress(target []string, output string, option CompressOption) error
	Extract(target string, output string, option ExtractOption) error
}

type BaseExtractOption struct {
	Password string
}

func (p *BaseExtractOption) GetPassword() string {
	return p.Password
}

type BaseCompressOption struct {
	Password string
}

func (o *BaseCompressOption) GetPassword() string {
	return o.Password
}

type ExtractFileOption struct {
	Input    string
	Output   string
	Password string
}

func ExtractArchive(option ExtractFileOption) error {
	var engine ArchiveEngine
	var extractOption ExtractOption
	if config.Instance.ArchiveEngine == config.ArchiveEngineWinRAR {
		engine = NewWinRAREngine(config.Instance.ArchiveCompress, config.Instance.ArchiveExtract)
		rarExtractOption := &WinRARExtractOption{}
		if len(option.Password) > 0 {
			rarExtractOption.Password = option.Password
		}
		extractOption = rarExtractOption
	}
	if config.Instance.ArchiveEngine == config.ArchiveEngineDefault {
		engine = &DefaultArchiveEngine{}
		extractOption = &DefaultArchiveExtractOption{
			BaseExtractOption{},
		}
	}
	err := engine.Extract(option.Input, option.Output, extractOption)
	return err
}
