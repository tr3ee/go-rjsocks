package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"syscall"

	"github.com/lxn/walk"
)

type releaseInfo struct {
	TagName string `json:"tag_name"`
}

var currentVersion = `v3.0.3b`

// CheckUpdate returns the latest tag name if there is a update
func CheckUpdate() (string, error) {
	latest := `https://api.github.com/repos/tr3ee/go-rjsocks/releases/latest`
	resp, err := http.Get(latest)
	if err != nil {
		return "", err
	}
	info := releaseInfo{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	if info.TagName != currentVersion {
		return info.TagName, nil
	}
	return "", nil
}

func AlertUpdate() {
	ver, err := CheckUpdate()
	if err != nil {
		nIcon.ShowError("无法检查更新", err.Error())
	} else {
		if len(ver) != 0 {
			nIcon.ShowInfo("RJSocks "+ver+" 更新", "当前版本：RJSocks "+currentVersion)
			ExecBackground("cmd", "/c", "start", "/b", "https://github.com/tr3ee/go-rjsocks/releases")
		} else {
			nIcon.ShowInfo("RJSocks "+currentVersion+" 当前已是最新版本", "点击关闭提示")
		}
	}
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
		AlertUpdate()
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
