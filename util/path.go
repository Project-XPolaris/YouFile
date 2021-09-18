package util

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ConvertPathWithOS(rawPath string) string {
	return filepath.FromSlash(rawPath)
}

func RenameDuplicateFilename(rawPath string) string {
	fileName := filepath.Base(rawPath)
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.ReplaceAll(fileName, ext, "")
	return filepath.Join(filepath.Dir(rawPath), fmt.Sprintf("%s_copy%s", nameWithoutExt, ext))
}
