package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Instance AppConfig
var Manager = viper.New()
var Logger = logrus.WithField("scope", "config")

type AppConfig struct {
	Addr        string
	FstabPath   string
	MountPoints []string
}

func LoadAppConfig() error {
	Manager.AddConfigPath("./")
	Manager.SetConfigName("config")
	Manager.SetConfigType("json")

	err := Manager.ReadInConfig()
	if err != nil {
		return err
	}
	Manager.SetDefault("addr", ":8300")
	Manager.SetDefault("fstab.path", "/etc/fstab")
	Manager.SetDefault("mountpoint", []string{})

	Instance.Addr = Manager.GetString("addr")
	Instance.FstabPath = Manager.GetString("fstab.path")
	Instance.MountPoints = Manager.GetStringSlice("mountpoint")
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
