package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/lxn/walk"
	rjsocks "github.com/tr3ee/go-rjsocks/core"
)

var (
	app        *walk.Application
	nIcon      *walk.NotifyIcon
	service    *rjsocks.Service
	mainWnd, _ = walk.NewMainWindow()
)

var (
	loginSubmitted bool
	iconSuccess, _ = walk.Resources.Icon("resources/rjsocks.ico")
	iconFailure, _ = walk.Resources.Icon("resources/stop.ico")
)

func init() {
	logfile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		walk.MsgBox(nil, "错误", "无法打开日志文件log.txt"+err.Error(), walk.MsgBoxIconError)
	}
	fileinfo, err := logfile.Stat()
	if err != nil {
		walk.MsgBox(nil, "错误", "无法获取log.txt文件信息"+err.Error(), walk.MsgBoxIconError)
	}
	if fileinfo.Size() > 10*1024 {
		logfile.Seek(0, 0)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(logfile)
	app = walk.App()
	app.SetProductName("rjsocks")
}

func main() {
	defer panicHandler()
	defer mainWnd.Close()
	defer appConfig.WriteBack()

	var err error
	nIcon, err = walk.NewNotifyIcon()
	if err != nil {
		panic(err)
	}
	defer nIcon.Dispose()
	setNotifyIcon()
	if !appConfig.AutoLogin && !runLoginFragment() {
		appExit(0)
		return
	}
	if err := nIcon.SetVisible(true); err != nil {
		panic(err)
	}
	allocService()
	go updateSrvStat()
	mainWnd.Run()
}

func updateSrvStat() {
	once := sync.Once{}
	lastState := rjsocks.SrvStat(-1)
	for {
		time.Sleep(1 * time.Second)
		if lastState != service.State {
			if err := nIcon.SetToolTip(service.State.String()); err != nil {
				break
			}
			lastState = service.State
			if lastState == rjsocks.SrvStatSuccess {
				nIcon.SetIcon(iconSuccess)
				once.Do(func() {
					nIcon.ShowMessage("RJSocks认证成功", "  GITHUB地址\nhttps://github.com/tr3ee/go-rjsocks")
				})
			} else if lastState == rjsocks.SrvStatFailure {
				nIcon.SetIcon(iconFailure)
				nIcon.ShowError("RJSocks认证失败", "当前设备未联网")
			}
		}
	}
}

func allocService() {
	var err error
	if service != nil {
		service.Close()
	}
	service, err = rjsocks.NewService(appConfig.Username, appConfig.Password, appConfig.Device, appConfig.Adapter)
	if err != nil {
		panic(err)
	}
	go service.Run()
}

func runLoginFragment() bool {
	loginSubmitted = false
	if err := LoginFragment(); err != nil {
		return false
	}
	return loginSubmitted
}

func setNotifyIcon() {
	nIcon.SetIcon(iconFailure)
	if err := nIcon.SetToolTip("RJSocks"); err != nil {
		panic(err)
	}

	enableAction := NewCheckableAction("启用网络(&E)", true)
	enableAction.Triggered().Attach(func() {
		enableAction.SetChecked(!enableAction.Checked())
		if enableAction.Checked() {
			service.Continue()
		} else {
			service.Stop()
		}
	})
	nIcon.ContextMenu().Actions().Add(enableAction)

	confAction := NewCheckableAction("允许自动登录", appConfig.AutoLogin)
	confAction.Triggered().Attach(func() {
		confAction.SetChecked(!confAction.Checked())
		appConfig.AutoLogin = confAction.Checked()
	})
	nIcon.ContextMenu().Actions().Add(confAction)

	loginAction := NewAction("断开连接(&O)")
	loginAction.Triggered().Attach(func() {
		if !runLoginFragment() {
			appExit(0)
		} else {
			allocService()
		}
	})
	nIcon.ContextMenu().Actions().Add(loginAction)

	renewAction := NewAction("刷新IP地址(&R)")
	renewAction.Triggered().Attach(func() {
		log.Println("刷新IP地址...")
		ExecBackground("ipconfig", "/renew", appConfig.Adapter)
		nIcon.ShowMessage("RJSocks 通知", "正在刷新IP地址...")
	})
	nIcon.ContextMenu().Actions().Add(renewAction)

	viewLogAction := NewAction("查看详细日志...")
	viewLogAction.Triggered().Attach(func() {
		go func() {
			logfullpath, err := filepath.Abs("./log.txt")
			if err != nil {
				log.Println(err)
				return
			}
			cmd := exec.Command("explorer.exe", `/select,`+logfullpath)
			if err := cmd.Run(); err != nil {
				log.Println(err)
			}
		}()
	})
	nIcon.ContextMenu().Actions().Add(viewLogAction)

	helpMenuAction, _ := nIcon.ContextMenu().Actions().AddMenu(NewHelpMenu())
	helpMenuAction.SetText("帮助(&H)")

	nIcon.ContextMenu().Actions().Add(walk.NewSeparatorAction())

	aboutAction := NewAction("关于 RJSocks(&A)")
	aboutAction.Triggered().Attach(func() {
		ExecBackground("cmd", "/c", "start", "/b", "https://github.com/tr3ee/go-rjsocks")
	})
	nIcon.ContextMenu().Actions().Add(aboutAction)

	exitAction := NewAction("退出(&X)")
	exitAction.Triggered().Attach(func() {
		nIcon.Dispose()
		appExit(0)
	})
	nIcon.ContextMenu().Actions().Add(exitAction)
}

func appExit(n int) {
	app.Exit(n)
	go func() {
		time.Sleep(6 * time.Second)
		os.Exit(-1)
	}()
}

func panicHandler() {
	if err := recover(); err != nil {
		walk.MsgBox(mainWnd, "错误", err.(error).Error(), walk.MsgBoxIconWarning)
		log.Fatal(err)
	}
}
