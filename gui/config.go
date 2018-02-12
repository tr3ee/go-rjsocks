package main

import (
	"log"
	"os"

	"github.com/astaxie/beego/config"
	"github.com/lxn/walk"
)

var appConfig *AppConfig

func init() {
	if fp, err := os.OpenFile("config.ini", os.O_CREATE|os.O_APPEND|os.O_RDONLY, 0666); err == nil {
		fp.Close()
	} else {
		walk.MsgBox(nil, "错误", "无法打开配置文件config.ini，查看日志log.txt获取详细信息", walk.MsgBoxIconError)
		log.Fatal("无法打开配置文件config.ini:" + err.Error())
	}
	appConfig = &AppConfig{}
	appConfig.ReadIn()
}

type AppConfig struct {
	configer                            config.Configer
	Username, Password, Device, Adapter string
	Remember, AutoLogin                 bool
}

func (c *AppConfig) ReadIn() {
	var err error
	c.configer, err = config.NewConfig("ini", "config.ini")
	if err != nil {
		log.Fatal(err)
	}
	c.Username = c.configer.DefaultString("username", "")
	c.Password = c.configer.DefaultString("password", "")
	c.Device = c.configer.DefaultString("device", "")
	c.Adapter = c.configer.DefaultString("adapter", "")
	c.Remember = c.configer.DefaultBool("Remember", true)
	c.AutoLogin = c.configer.DefaultBool("AutoLogin", false)
}

func (c *AppConfig) WriteBack() {
	c.configer.Set("username", c.Username)
	c.configer.Set("device", c.Device)
	c.configer.Set("adapter", c.Adapter)
	if c.Remember {
		c.configer.Set("password", c.Password)
		c.configer.Set("remember", "true")
		if c.AutoLogin {
			c.configer.Set("autologin", "true")
		} else {
			c.configer.Set("autologin", "false")
		}
	} else {
		c.configer.Set("password", "")
		c.configer.Set("remember", "false")
		c.configer.Set("autologin", "false")
	}
	c.configer.SaveConfigFile("config.ini")
}
