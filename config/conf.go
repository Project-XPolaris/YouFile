package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Instance AppConfig
var Manager = viper.New()
var Logger = logrus.WithField("scope", "config")

const (
	ArchiveEngineDefault = "Default"
	ArchiveEngineWinRAR  = "WinRAR"
)

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
}

func LoadAppConfig() error {
	Manager.AddConfigPath("./")
	Manager.AddConfigPath("../")
	Manager.SetConfigName("config")
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

	Instance.Addr = Manager.GetString("addr")
	Instance.FstabPath = Manager.GetString("fstab.path")
	Instance.MountPoints = Manager.GetStringSlice("mountpoint")
	Instance.YouPlusPath = Manager.GetBool("youplus.path")
	Instance.YouPlusUrl = Manager.GetString("youplus.url")
	Instance.YouPlusZFS = Manager.GetBool("youplus.zfs")
	Instance.YouPlusRPC = Manager.GetString("youplus.rpcurl")
	Instance.ArchiveEngine = Manager.GetString("archive.engine")
	Instance.ArchiveExtract = Manager.GetString("archive.extract")
	Instance.ArchiveCompress = Manager.GetString("archive.compress")
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
