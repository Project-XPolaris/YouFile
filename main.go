package main

import (
	"context"
	"fmt"
	srv "github.com/kardianos/service"
	entry "github.com/project-xpolaris/youplustoolkit/youplus/entity"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path/filepath"
	"youfile/api"
	"youfile/config"
	"youfile/database"
	"youfile/service"
	"youfile/util"
	"youfile/youplus"
)

var svcConfig *srv.Config
var Logger = logrus.WithField("scope", "main")

func initService() error {
	workPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	svcConfig = &srv.Config{
		Name:             "YouFileService",
		DisplayName:      "YouFile Service",
		WorkingDirectory: workPath,
		Arguments:        []string{"run"},
	}
	return nil
}
func Program() {
	err := config.LoadAppConfig()
	if err != nil {
		Logger.Fatal(err)
	}
	err = service.LoadFstab()
	if err != nil {
		Logger.Fatal(err)
	}
	err = database.ConnectToDatabase()
	if err != nil {
		Logger.Fatal(err)
	}
	if config.Instance.YouPlusPath {
		youplusLog := Logger.WithFields(logrus.Fields{
			"scope": "YouPlus",
			"url":   config.Instance.YouPlusUrl,
		})
		youplusLog.Info("check youplus service [checking]")
		err = youplus.InitClient()
		if err != nil {
			youplusLog.Fatal(err)
		}
		youplusLog.Info("check youplus service [pass]")
	}
	if len(config.Instance.YouPlusRPC) > 0 {
		youplusRPCLog := Logger.WithFields(logrus.Fields{
			"scope": "YouPlusRPC",
			"url":   config.Instance.YouPlusRPC,
		})
		youplusRPCLog.Info("check youplus rpc service [checking]")
		err = youplus.InitRPCClient()
		if err != nil {
			youplusRPCLog.Fatal(err)
		}
		youplusRPCLog.Info("check youplus rpc service [pass]")

	}
	// youplus entity
	if config.Instance.Entity.Enable {
		entityLog := Logger.WithFields(logrus.Fields{
			"scope": "YouplusEntity",
			"url":   config.Instance.YouPlusRPC,
		})
		entityLog.Info("register entity")
		youplus.InitEntity()

		err := youplus.DefaultEntry.Register()
		if err != nil {
			entityLog.Fatal(err.Error())
		}

		addrs, err := util.GetHostIpList()
		urls := make([]string, 0)
		for _, addr := range addrs {
			urls = append(urls, fmt.Sprintf("http://%s%s", addr, config.Instance.Addr))
		}
		if err != nil {
			entityLog.Fatal(err.Error())
		}
		err = youplus.DefaultEntry.UpdateExport(entry.EntityExport{Urls: urls, Extra: map[string]interface{}{}})
		if err != nil {
			entityLog.Fatal(err.Error())
		}

		err = youplus.DefaultEntry.StartHeartbeat(context.Background())
		if err != nil {
			entityLog.Fatal(err.Error())
		}
		entityLog.Info("success register entity")

	}
	api.RunApiService()
}

type program struct{}

func (p *program) Start(s srv.Service) error {
	go Program()
	return nil
}

func (p *program) Stop(s srv.Service) error {
	return nil
}

func InstallAsService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	s.Uninstall()

	err = s.Install()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("successful install service")
}

func UnInstall() {

	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	s.Uninstall()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("successful uninstall service")
}

func StartService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Start()
	if err != nil {
		logrus.Fatal(err)
	}
}
func StopService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Stop()
	if err != nil {
		logrus.Fatal(err)
	}
}
func RestartService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Restart()
	if err != nil {
		logrus.Fatal(err)
	}
}
func RunApp() {
	app := &cli.App{
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			&cli.Command{
				Name:  "service",
				Usage: "service manager",
				Subcommands: []*cli.Command{
					{
						Name:  "install",
						Usage: "install service",
						Action: func(context *cli.Context) error {
							InstallAsService()
							return nil
						},
					},
					{
						Name:  "uninstall",
						Usage: "uninstall service",
						Action: func(context *cli.Context) error {
							UnInstall()
							return nil
						},
					},
					{
						Name:  "start",
						Usage: "start service",
						Action: func(context *cli.Context) error {
							StartService()
							return nil
						},
					},
					{
						Name:  "stop",
						Usage: "stop service",
						Action: func(context *cli.Context) error {
							StopService()
							return nil
						},
					},
					{
						Name:  "restart",
						Usage: "restart service",
						Action: func(context *cli.Context) error {
							RestartService()
							return nil
						},
					},
				},
				Description: "YouFile service controller",
			},
			{
				Name:  "run",
				Usage: "run app",
				Action: func(context *cli.Context) error {
					Program()
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	err := initService()
	if err != nil {
		logrus.Fatal(err)
	}
	RunApp()
}
