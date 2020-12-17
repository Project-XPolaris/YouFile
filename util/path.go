package util

import (
	"path/filepath"
)

func ConvertPathWithOS(rawPath string) string {
	return filepath.FromSlash(rawPath)
}
