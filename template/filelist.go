package template

import (
	"os"
	"path/filepath"
)

type FileItem struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}
type FileListTemplate struct {
	Result []FileItem `json:"result"`
}

func NewFileListTemplate(result []os.FileInfo, parentPath string) *FileListTemplate {
	items := make([]FileItem, 0)
	for _, info := range result {
		item := FileItem{
			Name: info.Name(),
			Path: filepath.Join(parentPath, info.Name()),
			Size: info.Size(),
		}
		if info.IsDir() {
			item.Type = "Directory"
		} else {
			item.Type = "File"
		}
		items = append(items, item)
	}
	return &FileListTemplate{Result: items}
}
