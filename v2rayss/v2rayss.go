package v2rayss

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"v2rayss/core/parser"
	"v2rayss/core/vmess"

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

	coreServer *v2rayCore.Instance
	running    bool
	pingRound  int

	conf      *servInfo
	selectone int
	links     []*vmess.Link
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
	in, err := vmess.Vmess2Inbound(app.conf.addr, app.conf.proto, uint32(app.conf.port))
	if err != nil {
		return err
	}
	coreConf := makeV2rayConfig()
	coreConf.Inbound = []*v2rayCore.InboundHandlerConfig{in}
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

func (app *App) SaveConfInfo(addr, protocol string, port uint, subaddr string) error {
	app.mu.Lock()
	defer app.mu.Unlock()
	if addr != app.conf.addr || protocol != app.conf.proto || port != app.conf.port {
		if err := v2rayServeRunningHook(app.running, app.coreServer, func() error {
			// set v2ray in bound
			app.conf.addr = addr
			app.conf.proto = protocol
			app.conf.port = port
			in, err := vmess.Vmess2Inbound(app.conf.addr, app.conf.proto, uint32(app.conf.port))
			if err != nil {
				return err
			}
			return app.setCoreSeverDefaultIntbound(in)
		}); err != nil {
			return err
		}
	}
	if subaddr != app.conf.subAddr {
		app.conf.subAddr = subaddr
		return app.setSubAddr(subaddr)
	}
	return nil
}

func v2rayServeRunningHook(running bool, server *v2rayCore.Instance, f func() error) error {
	if running {
		if err := server.Close(); err != nil {
			return err
		}
	}
	if err := f(); err != nil {
		return err
	}
	if running {
		return server.Start()
	}
	return nil
}

func (app *App) setCoreSeverDefaultIntbound(in *v2rayCore.InboundHandlerConfig) error {
	inboundManager := app.coreServer.GetFeature(inbound.ManagerType()).(inbound.Manager)
	rawHandler, err := v2rayCore.CreateObject(app.coreServer, in)
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
	return inboundManager.AddHandler(context.Background(), handler)
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
	return app.loadSubAddr()
}

func (app *App) loadSubAddr() error {
	links, err := parser.Parse(app.conf.subAddr)
	if err != nil {
		return err
	}
	app.links = links
	return nil
}

func (app *App) PingLinks() error {
	app.mu.Lock()
	defer app.mu.Unlock()
	if app.links == nil {
		return nil
	}
	for i := range app.links {
		app.wg.Add(1)
		go func(i int) {
			defer app.wg.Done()
			if t, err := app.links[i].Ping(app.pingRound, "https://cloudflare.com/cdn-cgi/trace"); err == nil {
				app.links[i].PingTime = t
			} else {
				app.links[i].PingTime = vmess.NoPing
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
		return []string{}
	}
	res := make([]string, len(app.links))
	for i := range app.links {
		res[i] = fmt.Sprintf("%s - %s", app.links[i].PingTime, app.links[i].Ps)
	}
	return res
}

func (app *App) SelectLink(index int) error {
	app.mu.Lock()
	defer app.mu.Unlock()
	if index < 0 || index >= len(app.links) {
		return errors.New("Links index overflow")
	}
	return v2rayServeRunningHook(app.running, app.coreServer, func() error {
		// set v2ray out bound
		app.selectone = index
		out, err := vmess.Vmess2Outbound(app.links[app.selectone], true)
		if err != nil {
			return err
		}
		return app.setCoreSeverDefaultOutbound(out)
	})
}

func (app *App) setCoreSeverDefaultOutbound(out *v2rayCore.OutboundHandlerConfig) error {
	outboundManager := app.coreServer.GetFeature(outbound.ManagerType()).(outbound.Manager)
	rawHandler, err := v2rayCore.CreateObject(app.coreServer, out)
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
	return outboundManager.AddHandler(context.Background(), handler)
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

func makeV2rayConfig() *v2rayCore.Config {
	return &v2rayCore.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&applog.Config{
				ErrorLogType:  applog.LogType_Console,
				ErrorLogLevel: commlog.Severity_Error,
			}),
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
		},
		Outbound: make([]*v2rayCore.OutboundHandlerConfig, 0, 1),
		Inbound:  make([]*v2rayCore.InboundHandlerConfig, 0, 1),
	}
}
