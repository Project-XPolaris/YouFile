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
	CanCompress() bool
	CanExtract() bool
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
		if !engine.CanExtract() {
			engine = nil
		} else {
			rarExtractOption := &WinRARExtractOption{}
			if len(option.Password) > 0 {
				rarExtractOption.Password = option.Password
			}
			extractOption = rarExtractOption
		}
	}

	if config.Instance.ArchiveEngine == config.ArchiveEngineUnar {
		engine = NewUnarArchiveEngine()
		if !engine.CanExtract() {
			engine = nil
		} else {
			unarExtractOption := &UnarArchiveExtractOption{}
			if len(option.Password) > 0 {
				unarExtractOption.Password = option.Password
			}
			extractOption = unarExtractOption
		}
	}

	if engine == nil {
		engine = &DefaultArchiveEngine{}
		extractOption = &DefaultArchiveExtractOption{
			BaseExtractOption{},
		}
	}
	err := engine.Extract(option.Input, option.Output, extractOption)
	return err
}
