package main

import (
	"log"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/lxn/walk"
	"github.com/tr3ee/go-rjsocks"
)

var iconSuccess, _ = walk.Resources.Icon("resources/rjsocks.ico")
var iconFailure, _ = walk.Resources.Icon("resources/stop.ico")
var ni *walk.NotifyIcon

func UpdateStat() {
	once := sync.Once{}
	stat := rjsocks.SrvStat(-1)
	for {
		service.WaitStat()
		if stat != service.State {
			stat = service.State
			ni.SetToolTip(stat.String())
			if stat == rjsocks.SrvStatSuccess {
				ni.SetIcon(iconSuccess)
				once.Do(func() { ni.ShowMessage("RJSocks认证成功", "  GITHUB地址\nhttps://github.com/tr3ee/go-rjsocks") })
			} else if stat == rjsocks.SrvStatFailure {
				ni.SetIcon(iconFailure)
				ni.ShowError("RJSocks认证失败", "当前设备未联网")
			}
		}
	}
}

func LaunchNotifyIcon() error {
	mainWnd, err := walk.NewMainWindow()
	if err != nil {
		return err
	}
	defer mainWnd.Close()
	ni, err = walk.NewNotifyIcon()
	if err != nil {
		return err
	}
	defer ni.Dispose()
	// icon, err := walk.Resources.Icon("resources/rjsocks.ico")
	// if err != nil {
	// 	return err
	// }
	ni.SetIcon(iconFailure)
	if err := ni.SetToolTip("RJSocks"); err != nil {
		return err
	}

	// 设置菜单内容
	enableAction := NewCheckableAction("启用网络(&E)", true)
	enableAction.Triggered().Attach(func() {
		enableAction.SetChecked(!enableAction.Checked())
		if enableAction.Checked() {
			service.ContinueServe()
		} else {
			service.StopServe()
		}
	})
	ni.ContextMenu().Actions().Add(enableAction)

	confAction := NewCheckableAction("允许自动登录", appConfig.AutoLogin)
	confAction.Triggered().Attach(func() {
		confAction.SetChecked(!confAction.Checked())
		appConfig.AutoLogin = confAction.Checked()
	})
	ni.ContextMenu().Actions().Add(confAction)

	ni.ContextMenu().Actions().Add(walk.NewSeparatorAction())

	loginAction := NewAction("断开连接&&重新认证(&D)")
	loginAction.Triggered().Attach(func() {
		appConfig.AutoLogin = false
		ni.Dispose()
		// service.Close() cannot do it, the program crush otherwise
		main()
	})
	ni.ContextMenu().Actions().Add(loginAction)

	renewAction := NewAction("刷新IP地址(R)")
	renewAction.Triggered().Attach(func() {
		log.Println("刷新IP地址...")
		ExecBackground("ipconfig", "/renew", appConfig.Adapter)
		ni.ShowMessage("RJSocks 通知", "正在刷新IP地址...")
	})
	ni.ContextMenu().Actions().Add(renewAction)

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
	ni.ContextMenu().Actions().Add(viewLogAction)

	helpMenuAction, _ := ni.ContextMenu().Actions().AddMenu(NewHelpMenu())
	helpMenuAction.SetText("帮助(&H)")

	ni.ContextMenu().Actions().Add(walk.NewSeparatorAction())

	aboutAction := NewAction("关于 RJSocks(&A)")
	aboutAction.Triggered().Attach(func() {
		ExecBackground("cmd", "/c", "start", "/b", "https://github.com/tr3ee/go-rjsocks")
	})
	ni.ContextMenu().Actions().Add(aboutAction)
	exitAction := NewAction("退出(&X)")
	exitAction.Triggered().Attach(func() {
		ni.Dispose()
		app.Exit(0)
		// time.Sleep(2 * time.Second)
		// os.Exit(-1)
	})
	ni.ContextMenu().Actions().Add(exitAction)

	if err := ni.SetVisible(true); err != nil {
		return err
	}

	// if err := ni.ShowMessage("RJSocks已在后台运行", "ads"); err != nil {
	// 	return err
	// }
	go service.Serve()
	// 更新Icon状态
	go UpdateStat()
	mainWnd.Run()
	return nil
}

func NewAction(text string) *walk.Action {
	action := walk.NewAction()
	action.SetText(text)
	return action
}

func NewCheckableAction(text string, checked bool) *walk.Action {
	action := NewAction(text)
	action.SetCheckable(true)
	action.SetChecked(checked)
	return action
}

func NewHelpMenu() *walk.Menu {
	updateAction := NewAction("检查更新...")
	updateAction.Triggered().Attach(func() {
		ver, err := CheckUpdate()
		if err != nil {
			ni.ShowError("无法检查更新", err.Error())
		} else {
			if len(ver) != 0 {
				ni.ShowInfo("RJSocks "+ver+" 更新", "当前版本：RJSocks "+currentVersion)
				ExecBackground("cmd", "/c", "start", "/b", "https://github.com/tr3ee/go-rjsocks/releases")
			} else {
				ni.ShowInfo("RJSocks "+currentVersion+" 当前已是最新版本", "点击关闭提示")
			}
		}
	})
	usageAction := NewAction("使用帮助(&U)")
	usageAction.Triggered().Attach(func() {
		ExecBackground("cmd", "/c", "start", "/b", "https://github.com/tr3ee/go-rjsocks/wiki")
	})
	reportAction := NewAction("问题反馈(&R)")
	reportAction.Triggered().Attach(func() {
		ExecBackground("cmd", "/c", "start", "/b", "https://github.com/tr3ee/go-rjsocks/issues")
	})
	helpMenu, _ := walk.NewMenu()
	helpMenu.Actions().Add(updateAction)
	helpMenu.Actions().Add(usageAction)
	helpMenu.Actions().Add(reportAction)
	return helpMenu
}

func ExecBackground(name string, arg ...string) {
	go func() {
		cmd := exec.Command(name, arg...)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := cmd.Run(); err != nil {
			log.Println(err)
		}
	}()
}
