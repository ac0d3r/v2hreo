package v2rayss

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"v2rayss/core/parser"
	"v2rayss/core/vmess"

	"v2ray.com/core"
	v2rayCore "v2ray.com/core"
	"v2ray.com/core/app/dispatcher"
	applog "v2ray.com/core/app/log"
	"v2ray.com/core/app/proxyman"
	commlog "v2ray.com/core/common/log"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/features/inbound"
	"v2ray.com/core/features/outbound"
)

type servInfo struct {
	addr    string
	proto   string
	port    uint
	subAddr string // 订阅地址
}

type App struct {
	mu sync.Mutex
	wg sync.WaitGroup

	coreServer *core.Instance
	running    bool
	pingRound  int

	conf      *servInfo
	selectone int
	links     []*vmess.Link
	pings     []time.Duration
}

var (
	once sync.Once
	app  *App
)

func New() (*App, error) {
	var err error
	once.Do(func() {
		app = &App{}
		err = app.init()
	})
	return app, err
}

func (app *App) init() error {
	app.conf = &servInfo{
		addr:  "127.0.0.1",
		port:  1080,
		proto: "socks",
	}
	app.pingRound = 3
	// set v2ray inbound
	coreConf := makeV2rayConfig()
	in, err := vmess.Vmess2Inbound(app.conf.addr, app.conf.proto, uint32(app.conf.port))
	if err != nil {
		return err
	}
	coreConf.Inbound = []*core.InboundHandlerConfig{in}
	app.coreServer, err = v2rayCore.New(coreConf)
	if err != nil {
		app.coreServer = nil
		return err
	}
	return nil
}

func (app *App) Running() bool {
	return app.running
}

func (app *App) GetConfInfo() map[string]interface{} {
	return map[string]interface{}{
		"addr":    app.conf.addr,
		"port":    app.conf.port,
		"proto":   app.conf.proto,
		"subaddr": app.conf.subAddr,
	}
}

func (app *App) SaveConfInfo(addr string, port uint, subaddr string) error {
	app.mu.Lock()
	defer app.mu.Unlock()
	if addr != app.conf.addr || port != app.conf.port {
		if app.running {
			if err := app.coreServer.Close(); err != nil {
				return err
			}
			defer app.coreServer.Start()
		}
		app.conf.addr = addr
		app.conf.port = port
		in, err := vmess.Vmess2Inbound(app.conf.addr, app.conf.proto, uint32(app.conf.port))
		if err != nil {
			return err
		}
		app.setCoreSeverDefaultIntbound(in)
	}
	return app.setSubAddr(addr)
}

func (app *App) setCoreSeverDefaultIntbound(in *core.InboundHandlerConfig) error {
	inboundManager := app.coreServer.GetFeature(inbound.ManagerType()).(inbound.Manager)
	rawHandler, err := core.CreateObject(app.coreServer, in)
	if err != nil {
		return err
	}
	handler, ok := rawHandler.(inbound.Handler)
	if !ok {
		return errors.New("not an InboundHandler")
	}
	// remove old inbound
	if err := inboundManager.RemoveHandler(context.Background(), "proxy"); err != nil {
		return err
	}
	// add new in
	if err := inboundManager.AddHandler(context.Background(), handler); err != nil {
		return err
	}
	return nil
}

func (app *App) SetSubAddr(addr string) error {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.setSubAddr(addr)
}

func (app *App) setSubAddr(addr string) error {
	if addr != app.conf.subAddr {
		app.conf.subAddr = addr
		return app.loadSubAddr()
	}
	return nil
}

func (app *App) LoadSubAddr() error {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.LoadSubAddr()
}

func (app *App) loadSubAddr() error {
	links, err := parser.Parse(app.conf.subAddr)
	if err != nil {
		return err
	}
	app.links = links
	app.pings = make([]time.Duration, len(app.links))
	return nil
}

func (app *App) PingLinks() error {
	app.mu.Lock()
	defer app.mu.Unlock()
	if app.links == nil {
		return nil
	}
	app.makePings()
	for i := range app.links {
		app.wg.Add(1)
		go func(i int) {
			defer app.wg.Done()
			t, err := app.links[i].Ping(app.pingRound, "https://cloudflare.com/cdn-cgi/trace")
			if err == nil {
				app.pings[i] = t
			} else {
				app.pings[i] = vmess.NoPing
			}
		}(i)
	}
	app.wg.Wait()
	return nil
}

func (app *App) ListHosts() []string {
	app.mu.Lock()
	defer app.mu.Unlock()
	if app.links == nil {
		return nil
	}
	app.makePings()
	res := make([]string, len(app.links))
	for i := range app.links {
		res[i] = fmt.Sprintf("%s - %s", app.pings[i], app.links[i].Ps)
	}
	return res
}

func (app *App) makePings() {
	if app.pings == nil {
		app.pings = make([]time.Duration, len(app.links))
	}
}

func (app *App) SelectLink(index int) error {
	app.mu.Lock()
	defer app.mu.Unlock()
	if index < 0 && index >= len(app.links) {
		return errors.New("Index overflow")
	}
	app.selectone = index
	out, err := vmess.Vmess2Outbound(app.links[app.selectone], true)
	if err != nil {
		return err
	}
	if app.running {
		if err := app.coreServer.Close(); err != nil {
			return err
		}
		defer app.coreServer.Start()
	}
	app.setCoreSeverDefaultOutbound(out)
	return nil
}

func (app *App) setCoreSeverDefaultOutbound(out *core.OutboundHandlerConfig) error {
	outboundManager := app.coreServer.GetFeature(outbound.ManagerType()).(outbound.Manager)
	rawHandler, err := core.CreateObject(app.coreServer, out)
	if err != nil {
		return err
	}
	handler, ok := rawHandler.(outbound.Handler)
	if !ok {
		return errors.New("not an OutboundHandler")
	}
	// remove old out
	if h := outboundManager.GetHandler("proxy"); h != nil {
		if err := outboundManager.RemoveHandler(context.Background(), "proxy"); err != nil {
			return err
		}
	}
	// add new out
	if err := outboundManager.AddHandler(context.Background(), handler); err != nil {
		return err
	}
	return nil
}

func (app *App) Start() error {
	app.mu.Lock()
	defer app.mu.Unlock()
	if app.running {
		return nil
	}
	app.running = true
	return app.coreServer.Start()
}

func (app *App) Close() error {
	app.mu.Lock()
	defer app.mu.Unlock()
	if !app.running {
		return nil
	}
	app.running = false
	return app.coreServer.Close()
}

func makeV2rayConfig() *core.Config {
	return &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&applog.Config{
				ErrorLogType:  applog.LogType_Console,
				ErrorLogLevel: commlog.Severity_Error,
			}),
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
		},
		Outbound: make([]*core.OutboundHandlerConfig, 0, 1),
		Inbound:  make([]*core.InboundHandlerConfig, 0, 1),
	}
}
