package ui

import (
	"fmt"
	"v2rayss"
	"v2rayss/logs"
	"v2rayss/ui/icon"

	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
)

var (
	none        = ""
	hostsMenu   = []*systray.MenuItem{}
	infosMenu   = map[string]*systray.MenuItem{}
	checkedHost = -1
	app         = v2rayss.New()
)

func renderHosts(ShowServer *systray.MenuItem) {
	// init
	hs, index := app.HostList()
	mini := min(len(hs), len(hostsMenu))
	diff := len(hs) - len(hostsMenu)

	if len(hostsMenu) == 0 {
		for _, h := range hs {
			hostsMenu = append(hostsMenu, ShowServer.AddSubMenuItem(h, none))

		}
		goto End
	}
	for i := 0; i < mini; i++ {
		hostsMenu[i].Show()
		hostsMenu[i].SetTitle(hs[i])
	}
	if diff > 0 {
		for i := mini; i < len(hs); i++ {
			hostsMenu = append(hostsMenu, ShowServer.AddSubMenuItem(hs[i], none))
		}
	} else if diff < 0 {
		for i := mini; i < len(hostsMenu); i++ {
			hostsMenu[i].Hide()
		}
	}
End:
	checkHosts(index)
}

func checkHosts(index int) {
	if index != -1 && index < len(hostsMenu) {
		if checkedHost != -1 {
			hostsMenu[checkedHost].Uncheck()
		}
		checkedHost = index
		hostsMenu[checkedHost].Check()
	}
}

// OnExit App exit && cleanup
func OnExit() {
	err := app.Close()
	if err != nil {
		logs.Info(err)
	}
}

// OnReady App ready && ui init
func OnReady() {
	systray.SetIcon(icon.LogoOFF)
	systray.SetTitle("V2raySs")
	systray.SetTooltip("V2raySs")
	/*
		Menu Part
	*/
	Switch := systray.AddMenuItem("打开", "turn on/off")
	if app.CoreServStatus() == true {
		Switch.SetTitle("关闭")
	}
	Infos := systray.AddMenuItem("详情", none) // shwo infos
	for k, v := range app.StatusInfos() {
		switch k {
		case "listen":
			v = fmt.Sprintf("代理服务器：%s", v)
		case "port":
			v = fmt.Sprintf("代理端口： %s", v)
		case "protocol":
			v = fmt.Sprintf("代理协议： %s", v)
		case "subaddr":
			v = fmt.Sprintf("订阅地址： %s", v)
		}
		infosMenu[k] = Infos.AddSubMenuItem(v, none)
	}
	systray.AddSeparator()
	PasteSubAddr := systray.AddMenuItem("粘贴订阅地址", none)
	ShowServer := systray.AddMenuItem("服务器列表", none)
	PingServer := ShowServer.AddSubMenuItem("自动选择(ping)", none)
	SelectNextServer := ShowServer.AddSubMenuItem("next", none)
	SelectNextServer.SetIcon(icon.Next)
	renderHosts(ShowServer)
	systray.AddSeparator()
	Quit := systray.AddMenuItem("退出", none)

	go func() {
		for {
			select {
			case <-Switch.ClickedCh:
				if app.CoreServStatus() == false {
					err := app.TurnOn()
					if err != nil {
						logs.Info(err)
					} else {
						systray.SetIcon(icon.LogoON)
						Switch.SetTitle("关闭")
					}
				} else {
					err := app.TurnOff()
					if err != nil {
						logs.Info(err)
					} else {
						systray.SetIcon(icon.LogoOFF)
						Switch.SetTitle("打开")
					}
				}
			case <-PingServer.ClickedCh:
				go func() {
					app.Pings()
					renderHosts(ShowServer)
				}()
			case <-SelectNextServer.ClickedCh:
				if checkedHost != -1 {
					index, err := app.SelectHost(checkedHost + 1)
					if err != nil {
						logs.Info(err)
					} else {
						checkHosts(index)
					}
				}
			case <-PasteSubAddr.ClickedCh:
				addr, err := clipboard.ReadAll()
				if err != nil {
					logs.Info("parseSubaddr error: ", err)
				}
				infosMenu["subaddr"].SetTitle(fmt.Sprintf("订阅地址： %s", addr))
				err = app.UpdateSubAddr(addr)
				if err != nil {
					logs.Info(err)
				}
				renderHosts(ShowServer)
			case <-Quit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
