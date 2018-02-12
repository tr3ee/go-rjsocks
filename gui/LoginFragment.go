package main

import (
	"fmt"
	"log"
	"math/rand"
	_ "runtime/cgo"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/tr3ee/go-rjsocks/core"
)

func LoginFragment() error {
	LoginWnd, err := walk.NewMainWindow()
	if err != nil {
		return err
	}
	defer LoginWnd.Dispose()
	devs, err := rjsocks.ListNetworkDev()
	if err != nil {
		return err
	}
	adapters, err := rjsocks.ListNetworkAdapter()
	if err != nil {
		return err
	}

	var db *walk.DataBinder
	var checkBox1, checkBox2 *walk.CheckBox

	ButtonClickAction := func() {
		if err := db.Submit(); err != nil {
			walk.MsgBox(LoginWnd, "提示", "请填写所有必填项", walk.MsgBoxIconAsterisk)
			return
		}
		if err := LoginWnd.Close(); err != nil {
			log.Fatal(err)
		}
		loginSubmitted = true
	}
	KeyEnterAction := func(key walk.Key) {
		if key == walk.KeyReturn {
			ButtonClickAction()
		}
	}

	rand.Seed(time.Now().Unix())
	banner := fmt.Sprintf("banner-%d.jpg", rand.Intn(2)+1)

	mw := MainWindow{
		Title:   "校园网登录客户端 @tr3e",
		Size:    Size{320, 180},
		MaxSize: Size{360, 260},
		// Size:       Size{520, 390},
		// MaxSize:    Size{520, 390},
		Font:       Font{PointSize: 10},
		Background: SolidColorBrush{Color: walk.RGB(255, 255, 255)},
		Layout:     VBox{MarginsZero: true},
		AssignTo:   &LoginWnd,
		Visible:    false,
		DataBinder: DataBinder{
			AssignTo:   &db,
			DataSource: appConfig,
		},
		Children: []Widget{
			ImageView{
				Image:   "resources/" + banner,
				Mode:    ImageViewModeIdeal,
				MaxSize: Size{10, 0},
			},
			// VSeparator{},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSplitter{},
					Composite{
						Layout: Grid{Columns: 2},
						Children: []Widget{
							Label{Text: "用户名"},
							LineEdit{Text: Bind("Username", Regexp{"\\S+"}), OnKeyDown: KeyEnterAction},
							Label{Text: "密码"},
							LineEdit{Text: Bind("Password", Regexp{"\\S+"}), PasswordMode: true, OnKeyDown: KeyEnterAction},
							Label{Text: "网卡名称"},
							ComboBox{
								Value: Bind("Device", SelRequired{}),
								Model: devs,
							},
							Label{Text: "网络适配器"},
							ComboBox{
								Value: Bind("Adapter", SelRequired{}),
								Model: adapters,
							},
							Label{},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									CheckBox{AssignTo: &checkBox1, Text: "记住密码", Checked: Bind("Remember"), OnCheckStateChanged: func() {
										if !checkBox1.Checked() {
											checkBox2.SetChecked(false)
										}
									}},
									CheckBox{AssignTo: &checkBox2, Text: "自动登录", Checked: Bind("AutoLogin")},
								},
							},
							Label{},
							PushButton{
								Background: TransparentBrush{},
								Text:       "确定",
								OnClicked:  ButtonClickAction,
							},
						},
					},
					HSplitter{},
				},
			},
		},
	}
	if err := mw.Create(); err != nil {
		return err
	}
	LoginWnd.Show()
	LoginWnd.Run()
	return nil
}
