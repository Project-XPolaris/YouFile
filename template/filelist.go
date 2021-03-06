package template

import (
	"os"
	"path/filepath"
	"strings"
	"youfile/config"
	"youfile/service"
)

type FileItem struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Thumbnail string `json:"thumbnail,omitempty"`
}
type FileListTemplate struct {
	Result []FileItem `json:"result"`
	Sep    string     `json:"sep"`
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
			thumbnailName, _ := service.GetFileThumbnail(item.Path)
			item.Thumbnail = thumbnailName
		}
		items = append(items, item)
	}
	return &FileListTemplate{Result: items, Sep: string(filepath.Separator)}
}

func NewFileListTemplateFromTargetFile(result []service.TargetFile, src string) *FileListTemplate {

	items := make([]FileItem, 0)
	for _, targetFile := range result {
		targetPath := targetFile.Path
		if config.Instance.YouPlusPath {
			targetPath = strings.Replace(targetPath, src, targetFile.PathTrans, 1)
		}
		item := FileItem{
			Name: targetFile.Info.Name(),
			Path: targetPath,
			Size: targetFile.Info.Size(),
		}
		if targetFile.Info.IsDir() {
			item.Type = "Directory"
		} else {
			item.Type = "File"
		}
		items = append(items, item)
	}
	return &FileListTemplate{Result: items}
}
