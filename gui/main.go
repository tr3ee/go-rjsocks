package main

import (
	"log"
	"os"

	"github.com/lxn/walk"
	"github.com/tr3ee/go-rjsocks"
)

var (
	app            *walk.Application
	appConfig      *AppConfig
	LoginSubmitted bool
	service        *rjsocks.Service
)

func init() {
	if fp, err := os.OpenFile("config.ini", os.O_CREATE|os.O_APPEND|os.O_RDONLY, 0666); err == nil {
		fp.Close()
	} else {
		walk.MsgBox(nil, "错误", "无法打开配置文件config.ini，查看日志log.txt获取详细信息", walk.MsgBoxIconError)
		log.Fatal("无法打开配置文件config.ini:" + err.Error())
	}
	logfile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		walk.MsgBox(nil, "错误", "无法打开日志文件log.txt"+err.Error(), walk.MsgBoxIconError)
		log.Fatal("无法打开日志文件log.txt:" + err.Error())
	}
	fileinfo, err := logfile.Stat()
	if err == nil {
		if fileinfo.Size() > 10*1024*1024 {
			logfile.Seek(0, 0)
		}
	}
	log.SetOutput(logfile)
	app = walk.App()
	app.SetProductName("rjsocks")
	appConfig = &AppConfig{}
	appConfig.ReadIn()
}

func main() {
	var err error
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()
	LoginSubmitted = false
	if err := LaunchLoginWindow(); err != nil {
		log.Fatal(err)
	}
	if LoginSubmitted {
		service, err = rjsocks.NewService(appConfig.Username, appConfig.Password, appConfig.Device, appConfig.Adapter)
		if err != nil {
			log.Fatal(err)
		}
		defer service.Close()
		defer appConfig.WriteBack()
		if err := LaunchNotifyIcon(); err != nil {
			log.Fatal(err)
		}
	}
}
