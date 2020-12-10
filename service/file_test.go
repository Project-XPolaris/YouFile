package service

import (
	"fmt"
	"testing"
	"time"
)

//func TestFile(t *testing.T) {
//	user, err := user.Current()
//	if err != nil {
//		t.Error(err)
//	}
//	file, err := AppFs.Open(user.HomeDir)
//	if err != nil {
//		t.Error(err)
//	}
//	items, err := file.Readdir(0)
//	if err != nil {
//		t.Error(err)
//	}
//	for _, item := range items {
//		fmt.Println(item.Name())
//	}
//}
//
//func TestCopyReaderText(t *testing.T) {
//	src, err := AppFs.Open("C:\\Users\\Takay\\Desktop\\New folder\\ubuntu-20.04.1-live-server-amd64.iso")
//	if err != nil {
//		t.Error(err)
//	}
//	reader := util.NewCounterReader(src)
//	dst, err := AppFs.OpenFile("C:\\Users\\Takay\\Desktop\\New folder\\1ubuntu-20.04.1-live-server-amd64.iso", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775)
//	if err != nil {
//		t.Error(err)
//	}
//	go func() {
//		<-time.After(2 * time.Second)
//		reader.StopChan <- struct{}{}
//	}()
//	_,err  = io.Copy(dst,reader)
//	fmt.Println("complete")
//	if err != nil  {
//		t.Error(err)
//	}
//}

func TestStopCopyFileTask(t *testing.T) {
	var task *Task
	go func() {
		<-time.After(6 * time.Second)
		DefaultTask.Lock()
		if task != nil {
			task.InterruptChan <- struct{}{}
		}
		DefaultTask.Unlock()
	}()
	go func() {
		for {
			<-time.After(1 * time.Second)
			DefaultTask.Lock()
			fmt.Println(task.Status)
			DefaultTask.Unlock()
		}
	}()
	task = DefaultTask.NewCopyFileTask("C:\\Users\\Takay\\Desktop\\New folder\\youmusic2", "C:\\Users\\Takay\\Desktop\\New folder\\youmusic3")
	<-time.After(10 * time.Second)
}
