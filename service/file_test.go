package service

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
