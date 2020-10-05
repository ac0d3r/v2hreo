package v2rayss

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
	"v2rayss/logs"
	"v2rayss/utils"
	"v2rayss/vmess"

	"v2ray.com/core"
)

// App of V2raySs
type App struct {
	coreServer *core.Instance
	lock       *sync.Mutex
	// coreServStatus v2ray-core serv status
	// true => started/false => closed
	coreStatus bool
	//subscribe address
	subAddr string
	subf    string
	// serverList List of all available servers
	serverList []*vmess.Host
	pings      []time.Duration
	pingRound  int
	// inbound
	listen   string
	protocol string
	port     uint32
	inbound  *core.InboundHandlerConfig
}

var (
	once sync.Once
	app  *App
	ctx  = context.Background()
)

// New return an instance of v2rayss app
func New() *App {
	once.Do(func() {
		app = &App{listen: "127.0.0.1", protocol: "socks", port: 1080}
		subf, err := utils.CheckAppDir(utils.UserDir, ".sub")
		if err != nil {
			logs.Fatalln(err)
		}
		app.subf = subf
		app.pingRound = 3
		app.lock = new(sync.Mutex)
		app.loadSubAddr()
		app.loadServerList()
		app.makeV2rayCore()
		app.coreStatus = false
	})
	return app
}

func (s *App) makeV2rayCore() {
	// makeV2rayCore init v2ray-core Inbound
	if s.inbound == nil {
		inbound, err := vmess.Vmess2Inbound(s.listen, s.protocol, s.port)
		if err != nil {
			logs.Fatalln(err)
		} else {
			s.inbound = inbound
		}
	}
	v2core, err := vmess.StartV2Ray(false, s.inbound, nil)
	if err != nil {
		logs.Fatalln(err)
		return
	}
	if app.coreStatus {
		s.coreServer.Close()
		s.coreServer = v2core
		s.coreServer.Start()
	}
	s.coreServer = v2core
}

//CoreServStatus v2ray-core server status
func (s *App) CoreServStatus() bool {
	return s.coreStatus
}

//StatusInfos app status info
func (s *App) StatusInfos() map[string]string {
	return map[string]string{
		"listen":   s.listen,
		"port":     fmt.Sprintf("%d", s.port),
		"protocol": s.protocol,
		"subaddr":  s.subAddr,
	}
}

// TurnOn turn on v2ray-core serv
func (s *App) TurnOn() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.coreServer == nil {
		return errors.New("Not v2ray-core instance")
	}
	if s.coreStatus == true {
		return errors.New("v2ray-core started")
	}
	err := s.coreServer.Start()
	if err != nil {
		return err
	}
	s.coreStatus = true
	return nil
}

// TurnOff turn off v2ray-core serv
func (s *App) TurnOff() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.coreServer == nil {
		return errors.New("Not v2ray-core instance")
	}
	if s.coreStatus == false {
		return errors.New("v2ray-core stopped")
	}
	err := s.coreServer.Close()
	if err != nil {
		return err
	}
	s.coreStatus = false
	return nil
}

/*
	Subscribe Address part
*/

func (s *App) loadSubAddr() {
	f, err := ioutil.ReadFile(s.subf)
	if err != nil {
		logs.Info("loadSubAddr fail", err)
	}
	if string(f) != "" {
		s.subAddr = string(f)
	}
}

func (s *App) storeSubAddr() {
	if s.subAddr == "" {
		return
	}
	err := ioutil.WriteFile(s.subf, []byte(s.subAddr), 0666)
	if err != nil {
		logs.Info("storeSubAddr fail", err)
	}
}

// UpdateSubAddr update subscribe address
func (s *App) loadServerList() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	hosts, err := vmess.ParseSubscription(s.subAddr)
	if err != nil {
		logs.Info(err)
		return err
	}
	s.serverList = hosts
	return nil
}

// UpdateSubAddr update subscribe address
func (s *App) UpdateSubAddr(addr string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	hosts, err := vmess.ParseSubscription(addr)
	if err != nil {
		logs.Info(err)
		return err
	}
	s.subAddr = addr
	s.serverList = hosts
	return nil
}

/*
	Servers part
*/

// HostList return app server `name-ping` list
func (s *App) HostList() ([]string, int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	hosts := []string{}
	for _, host := range s.serverList {
		hosts = append(hosts, host.Ps)
	}
	for i, ts := range s.pings {
		hosts[i] = fmt.Sprintf("%s %s", ts, hosts[i])
	}
	index := s.autoSelectServer()
	if index != -1 {
		out, err := vmess.Vmess2Outbound(s.serverList[index], true)
		if err != nil {
			logs.Info(err)
		} else {
			s.checkCoreServerOutbound(out)
		}
	}
	return hosts, index
}

// SelectHost free choice of servers
func (s *App) SelectHost(index int) (int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if index < 0 || index >= len(s.serverList) {
		index = 0
	}
	out, err := vmess.Vmess2Outbound(s.serverList[index], true)
	if err != nil {
		logs.Info(err)
		return -1, err
	}
	s.checkCoreServerOutbound(out)
	return index, nil
}

func (s *App) checkCoreServerOutbound(out *core.OutboundHandlerConfig) {
	s.makeV2rayCore()
	core.AddOutboundHandler(s.coreServer, out)
}

// Pings select server index number in `s.serverList`
func (s *App) Pings() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.pings) > 0 || len(s.serverList) == 0 {
		s.pings = []time.Duration{}
		if len(s.serverList) == 0 {
			return
		}
	}

	//test destination url (vmess ping only)
	pingData := make(chan time.Duration, len(s.serverList))
	dst := "https://cloudflare.com/cdn-cgi/trace"
	for _, host := range s.serverList {
		go func(host *vmess.Host, round int, dst string) {
			ctx, cannel := context.WithTimeout(ctx, 3*time.Duration(round)*time.Second)
			defer cannel()
			t, err := vmess.Ping(ctx, host, round, dst)
			if err != nil {
				logs.Info(err)
			}
			pingData <- t
		}(host, s.pingRound, dst)
	}

	count := 0
LOOP:
	for {
		select {
		case t := <-pingData:
			s.pings = append(s.pings, t)
			count++
			if count >= len(s.serverList) {
				break LOOP
			}
		}
	}
}

func (s *App) autoSelectServer() int {
	if len(s.pings) == 0 {
		return -1
	}
	min := -1
	var fast time.Duration
	for index, td := range s.pings {
		if td == vmess.NoPing {
			continue
		}
		// init fast
		if fast == 0 {
			fast = td
			min = index
			continue
		}
		// compare ping times
		if td < fast {
			fast = td
			min = index
		}
	}
	return min
}

// Close v2rayss app
func (s *App) Close() error {
	err := s.TurnOff()
	s.storeSubAddr()
	return err
}
