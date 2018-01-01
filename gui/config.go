package main

import (
	"log"

	"github.com/astaxie/beego/config"
)

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
