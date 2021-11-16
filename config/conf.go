package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

var Instance AppConfig
var Manager = viper.New()
var Logger = logrus.WithField("scope", "config")

const (
	ArchiveEngineDefault = "Default"
	ArchiveEngineWinRAR  = "WinRAR"
	ArchiveEngineUnar    = "Unar"
)

type EntityConfig struct {
	Enable  bool
	Name    string
	Version int64
}
type YouLogConfig struct {
	Remote      bool
	Addr        string
	Retry       int
	Application string
	Instance    string
}
type YouLinkConfig struct {
	Enable     bool
	Url        string
	ServiceUrl string
}
type AppConfig struct {
	Addr            string
	FstabPath       string
	MountPoints     []string
	YouPlusPath     bool
	YouPlusAuth     bool
	YouPlusUrl      string
	YouPlusZFS      bool
	YouPlusRPC      string
	ArchiveEngine   string
	ArchiveExtract  string
	ArchiveCompress string
	Thumbnails      bool
	Entity          EntityConfig
	YouLog          YouLogConfig
	Remote          RemoteConfig
	YouLink         YouLinkConfig
}
type RemoteServerConfig struct {
	Enable bool
	Addr   string
}
type RemoteClientConfig struct {
	Enable bool
	Addrs  []string
}
type RemoteConfig struct {
	Server RemoteServerConfig
	Client RemoteClientConfig
}

func LoadAppConfig(configPath string) error {
	if len(configPath) != 0 {
		Manager.AddConfigPath(filepath.Dir(configPath))
		configFile := filepath.Base(configPath)
		configFile = strings.ReplaceAll(configFile, filepath.Ext(configFile), "")
		Manager.SetConfigName(configFile)

	} else {
		Manager.AddConfigPath("./")
		Manager.SetConfigName("config")
	}

	Manager.SetConfigType("json")

	err := Manager.ReadInConfig()
	if err != nil {
		return err
	}
	Manager.SetDefault("addr", ":8300")
	Manager.SetDefault("fstab.path", "/etc/fstab")
	Manager.SetDefault("mountpoint", []string{})
	Manager.SetDefault("youplus.path", false)
	Manager.SetDefault("youplus.url", "http://localhost:8999")
	Manager.SetDefault("youplus.rpcurl", "")
	Manager.SetDefault("youplus.zfs", false)
	Manager.SetDefault("archive.engine", ArchiveEngineDefault)
	Manager.SetDefault("thumbnails", false)
	Manager.SetDefault("youlog.addr", "localhost:50052")
	Manager.SetDefault("youlog.remote", false)
	Manager.SetDefault("youlog.retry", 3000)
	Manager.SetDefault("youlog.app", "YouFileCoreService")

	Manager.SetDefault("remote.server.enable", false)
	Manager.SetDefault("remote.server.addr", "localhost:50060")
	Manager.SetDefault("remote.client.enable", false)
	Manager.SetDefault("remote.client.addrs", []string{})
	Instance.Addr = Manager.GetString("addr")
	Instance.FstabPath = Manager.GetString("fstab.path")
	Instance.MountPoints = Manager.GetStringSlice("mountpoint")
	Instance.YouPlusPath = Manager.GetBool("youplus.path")
	Instance.YouPlusUrl = Manager.GetString("youplus.url")
	Instance.YouPlusZFS = Manager.GetBool("youplus.zfs")
	Instance.YouPlusAuth = Manager.GetBool("youplus.auth")
	Instance.YouPlusRPC = Manager.GetString("youplus.rpcurl")
	Instance.ArchiveEngine = Manager.GetString("archive.engine")
	Instance.ArchiveExtract = Manager.GetString("archive.extract")
	Instance.ArchiveCompress = Manager.GetString("archive.compress")
	Instance.Thumbnails = Manager.GetBool("thumbnails")
	Instance.Entity = EntityConfig{
		Enable:  Manager.GetBool("youplus.entity.enable"),
		Name:    Manager.GetString("youplus.entity.name"),
		Version: Manager.GetInt64("youplus.entity.version"),
	}
	Instance.YouLog = YouLogConfig{
		Remote:      Manager.GetBool("youlog.remote"),
		Addr:        Manager.GetString("youlog.addr"),
		Retry:       Manager.GetInt("youlog.retry"),
		Application: Manager.GetString("youlog.app"),
		Instance:    Manager.GetString("youlog.instance"),
	}
	Instance.Remote = RemoteConfig{
		Server: RemoteServerConfig{
			Enable: Manager.GetBool("remote.server.enable"),
			Addr:   Manager.GetString("remote.server.addr"),
		},
		Client: RemoteClientConfig{
			Enable: Manager.GetBool("remote.client.enable"),
			Addrs:  Manager.GetStringSlice("remote.client.addrs"),
		},
	}
	Instance.YouLink = YouLinkConfig{
		Enable:     Manager.GetBool("youlink.enable"),
		Url:        Manager.GetString("youlink.url"),
		ServiceUrl: Manager.GetString("youlink.service"),
	}
	return nil
}

func SaveConfig() error {
	Logger.Info("save info")
	return Manager.WriteConfig()
}

func SaveMounts() error {
	Manager.Set("mountpoint", Instance.MountPoints)
	return SaveConfig()
}
