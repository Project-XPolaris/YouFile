package service

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/nfnt/resize"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"youfile/database"
)

func ReadDir(readPath string) ([]os.FileInfo, error) {
	file, err := AppFs.Open(readPath)
	if err != nil {
		return nil, err
	}
	items, err := file.Readdir(0)
	return items, nil
}
func Copy(src string, dest string, notifier *CopyFileNotifier, onDuplicate string) error {
	srcStat, err := AppFs.Stat(src)
	if err != nil {
		return err
	}
	if srcStat.IsDir() {
		err = CopyDir(src, dest, notifier, onDuplicate)
		if err != nil {
			return err
		}
	} else {
		err = CopyFile(src, dest, notifier, onDuplicate)
		if err != nil {
			return err
		}
	}
	return nil
}
func Move(src string, dest string, notifier *MoveFileNotifier, onDuplicate string) error {
	srcStat, err := AppFs.Stat(src)
	if err != nil {
		return err
	}
	if srcStat.IsDir() {
		err = MoveDir(src, dest, notifier, onDuplicate)
		if err != nil {
			return err
		}
	} else {
		err = MoveFile(src, dest, notifier, onDuplicate)
		if err != nil {
			return err
		}
	}
	return nil
}
func Rename(oldName string, newName string) error {
	return AppFs.Rename(oldName, newName)
}
func DeleteFile(target string) error {
	return AppFs.RemoveAll(target)
}

func Chmod(target string, perm int) error {
	return AppFs.Chmod(target, os.FileMode(perm))
}

type CopyAnalyzeResult struct {
	FileCount int
	DirCount  int
	TotalSize int64
}

func analyzeSource(src string) (*CopyAnalyzeResult, error) {
	sourceStat, err := AppFs.Stat(src)
	if err != nil {
		return nil, err
	}
	if sourceStat.IsDir() {
		fileCounter := 0
		dirCount := 0
		var totalSize int64 = 0
		err := afero.Walk(AppFs, src, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				dirCount += 1
			} else {
				fileCounter += 1
			}
			totalSize += info.Size()
			return nil
		})
		return &CopyAnalyzeResult{FileCount: fileCounter, TotalSize: totalSize}, err
	} else {
		return &CopyAnalyzeResult{
			FileCount: 1,
			DirCount:  0,
			TotalSize: sourceStat.Size(),
		}, nil
	}
}

var (
	DeleteInterrupt = errors.New("delete interrupt")
)

type DeleteNotifier struct {
	DeleteChan     chan string
	DeleteDoneChan chan string
	Info           *CopyAnalyzeResult
	StopFlag       bool
}

func Delete(src string, notifier *DeleteNotifier) error {
	srcStat, err := AppFs.Stat(src)
	if err != nil {
		return err
	}
	if srcStat.IsDir() {
		err = afero.Walk(AppFs, src, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				if notifier != nil {
					notifier.DeleteChan <- src
				}
				err := AppFs.Remove(path)
				//fmt.Println(path)
				if err != nil {
					return err
				}
				if notifier != nil {
					notifier.DeleteDoneChan <- src
				}
			}
			if notifier != nil && notifier.StopFlag {
				fmt.Println("try to stop delete")
				return DeleteInterrupt
			}
			return nil
		})
		err = AppFs.RemoveAll(src)
	} else {
		if notifier != nil {
			notifier.DeleteChan <- src
		}
		err = AppFs.Remove(src)
		if notifier != nil {
			notifier.DeleteDoneChan <- src
		}
		if notifier != nil && notifier.StopFlag {
			return DeleteInterrupt
		}
	}
	return err
}

func NewDirectory(dirPath string, perm int) error {
	err := AppFs.MkdirAll(dirPath, os.FileMode(perm))
	return err
}

func NewTextFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

func WriteTextFile(path string, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

func ReadFileAsString(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	text := string(content)
	return text, nil
}
func GetFileCheckSum(path string) (string, error) {
	file, err := os.Open(path)

	if err != nil {
		return "", err
	}

	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
func createThumbnailImage(path string, saveName string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	ext := filepath.Ext(path)
	ext = strings.ToLower(ext)

	var decoder func(r io.Reader) (image.Image, error)
	if ext == ".jpg" || ext == ".jpeg" {
		decoder = jpeg.Decode
	}
	if ext == ".png" {
		decoder = png.Decode
	}
	if decoder == nil {
		return nil
	}

	img, err := decoder(file)
	if err != nil {
		return err
	}

	file.Close()

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(240, 0, img, resize.Lanczos3)

	out, err := os.Create(filepath.Join("./thumbnails", saveName))
	if err != nil {
		return err
	}
	defer out.Close()

	// write new image to file
	if ext == ".jpg" || ext == ".jpeg" {
		err = jpeg.Encode(out, m, nil)
		if err != nil {
			return err
		}
	}
	if ext == ".png" {
		err = png.Encode(out, m)
		if err != nil {
			return err
		}
	}
	return nil
}

var AllowGenerateThumbnailImageExtensions = []string{
	".jpg", ".png", ".jpeg",
}

func GenerateImageThumbnail(rootPath string, onComplete func()) error {
	err := os.MkdirAll("./thumbnails", os.ModePerm)
	if err != nil {
		return err
	}
	items, err := afero.ReadDir(AppFs, rootPath)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		itemPath := filepath.Join(rootPath, item.Name())
		if !linq.From(AllowGenerateThumbnailImageExtensions).Contains(strings.ToLower(filepath.Ext(itemPath))) {
			continue
		}
		sum, err := GetFileCheckSum(itemPath)
		if err != nil {
			logrus.Error(err)
			continue
		}
		isExist, _ := afero.Exists(AppFs, filepath.Join("./thumbnails", fmt.Sprintf("%s%s", sum, filepath.Ext(item.Name()))))
		if isExist {
			logrus.Info("skip thumbnail exist")
			continue
		}
		thumbnailFilename := fmt.Sprintf("%s%s", sum, filepath.Ext(item.Name()))
		err = createThumbnailImage(itemPath, thumbnailFilename)
		if err != nil {
			logrus.Error(err)
			continue
		}
		database.Instance.Create(&database.Thumbnail{Path: itemPath, Checksum: sum})
	}
	onComplete()
	return nil
}

func GetFileThumbnail(path string) (string, error) {
	if !linq.From(AllowGenerateThumbnailImageExtensions).Contains(strings.ToLower(filepath.Ext(path))) {
		return "", nil
	}
	sum, err := GetFileCheckSum(path)
	if err != nil {
		return "", err
	}
	thumbnailFilename := fmt.Sprintf("%s%s", sum, filepath.Ext(path))
	isExist, err := afero.Exists(AppFs, filepath.Join("./thumbnails", thumbnailFilename))
	if err != nil {
		return "", err
	}
	if isExist {
		return thumbnailFilename, nil
	}
	return "", nil
}

type ClearThumbnailOption struct {
	All bool `hsource:"query" hname:"all"`
}

func ClearThumbnail(option ClearThumbnailOption) error {
	var err error
	if option.All {
		err = os.RemoveAll("./thumbnails")
		if err != nil {
			return err
		}
		err = database.Instance.Model(&database.Thumbnail{}).Unscoped().Where("id != ?", -1).Delete(&database.Thumbnail{}).Error
		if err != nil {
			return err
		}
		return nil
	}
	var thumbnails []database.Thumbnail
	err = database.Instance.Find(&thumbnails).Error
	if err != nil {
		return err
	}
	for _, thumbnail := range thumbnails {
		isExist, _ := afero.Exists(AppFs, thumbnail.Path)
		if !isExist {
			os.Remove(filepath.Join("./thumbnails", fmt.Sprintf("%s%s", thumbnail.Checksum, filepath.Ext(thumbnail.Path))))
			database.Instance.Model(&database.Thumbnail{}).Unscoped().Delete(thumbnail)
		}
	}
	return nil
}
