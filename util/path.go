package util

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
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

func GenerateRPCTimeoutContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	return ctx
}
