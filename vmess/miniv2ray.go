package vmess

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"v2ray.com/core"
	"v2ray.com/core/app/dispatcher"
	applog "v2ray.com/core/app/log"
	"v2ray.com/core/app/proxyman"
	commlog "v2ray.com/core/common/log"
	v2net "v2ray.com/core/common/net"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/infra/conf"
)

// Vmess2Inbound  v2ray-core set intbound
func Vmess2Inbound(listen, protocol string, port uint32) (*core.InboundHandlerConfig, error) {
	in := &conf.InboundDetourConfig{}
	v := conf.Address{}
	v.Address = v2net.ParseAddress(listen)
	in.ListenOn = &v
	in.Protocol = protocol
	in.PortRange = &conf.PortRange{From: port, To: port}
	inset := json.RawMessage([]byte(fmt.Sprintf(`{"udp": true}`)))
	in.Settings = &inset
	return in.Build()
}

// Vmess2Outbound  v2ray-core outbound host
func Vmess2Outbound(v *Host, usemux bool) (*core.OutboundHandlerConfig, error) {
	out := &conf.OutboundDetourConfig{}
	out.Tag = "proxy"
	out.Protocol = "vmess"
	out.MuxSettings = &conf.MuxConfig{}
	if usemux {
		out.MuxSettings.Enabled = true
		out.MuxSettings.Concurrency = 8
	}

	p := conf.TransportProtocol(v.Network)
	s := &conf.StreamConfig{
		Network:  &p,
		Security: v.TLS,
	}

	switch v.Network {
	case "tcp":
		s.TCPSettings = &conf.TCPConfig{}
		if v.Type == "" || v.Type == "none" {
			s.TCPSettings.HeaderConfig = json.RawMessage([]byte(`{ "type": "none" }`))
		} else {
			pathb, _ := json.Marshal(strings.Split(v.Path, ","))
			hostb, _ := json.Marshal(strings.Split(v.Host, ","))
			s.TCPSettings.HeaderConfig = json.RawMessage([]byte(fmt.Sprintf(`
			{
				"type": "http",
				"request": {
					"path": %s,
					"headers": {
						"Host": %s
					}
				}
			}
			`, string(pathb), string(hostb))))
		}
	case "kcp":
		s.KCPSettings = &conf.KCPConfig{}
		s.KCPSettings.HeaderConfig = json.RawMessage([]byte(fmt.Sprintf(`{ "type": "%s" }`, v.Type)))
	case "ws":
		s.WSSettings = &conf.WebSocketConfig{}
		s.WSSettings.Path = v.Path
		s.WSSettings.Headers = map[string]string{
			"Host": v.Host,
		}
	case "h2", "http":
		s.HTTPSettings = &conf.HTTPConfig{
			Path: v.Path,
		}
		if v.Host != "" {
			h := conf.StringList(strings.Split(v.Host, ","))
			s.HTTPSettings.Host = &h
		}
	}

	if v.TLS == "tls" {
		s.TLSSettings = &conf.TLSConfig{
			Insecure: true,
		}
		if v.Host != "" {
			s.TLSSettings.ServerName = v.Host
		}
	}

	out.StreamSetting = s
	oset := json.RawMessage([]byte(fmt.Sprintf(`{
  "vnext": [
    {
      "address": "%s",
      "port": %v,
      "users": [
        {
          "id": "%s",
          "alterId": %v,
          "security": "auto"
        }
      ]
    }
  ]
}`, v.Address, v.Port, v.ID, v.Aid)))
	out.Settings = &oset
	return out.Build()
}

// StartV2Ray start v2ray-core
func StartV2Ray(verbose bool, inbound *core.InboundHandlerConfig, outbound *core.OutboundHandlerConfig) (*core.Instance, error) {
	loglevel := commlog.Severity_Error
	if verbose {
		loglevel = commlog.Severity_Debug
	}
	config := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&applog.Config{
				ErrorLogType:  applog.LogType_Console,
				ErrorLogLevel: loglevel,
			}),
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
		},
	}

	// commlog.RegisterHandler(commlog.NewLogger(commlog.CreateStderrLogWriter()))
	if outbound != nil {
		config.Outbound = []*core.OutboundHandlerConfig{outbound}
	}
	if inbound != nil {
		config.Inbound = []*core.InboundHandlerConfig{inbound}
	}
	server, err := core.New(config)
	if err != nil {
		return nil, err
	}
	return server, nil
}

func measureDelay(inst *core.Instance, timeout time.Duration, dest string) (time.Duration, error) {
	start := time.Now()
	code, _, err := CoreHTTPRequest(inst, timeout, "GET", dest)
	if err != nil {
		return -1, err
	}
	if code > 399 {
		return -1, fmt.Errorf("status incorrect (>= 400): %d", code)
	}
	return time.Since(start), nil
}

// CoreHTTPClient v2ray core http client
func CoreHTTPClient(inst *core.Instance, timeout time.Duration) (*http.Client, error) {

	if inst == nil {
		return nil, errors.New("core instance nil")
	}

	tr := &http.Transport{
		DisableKeepAlives: true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dest, err := v2net.ParseDestination(fmt.Sprintf("%s:%s", network, addr))
			if err != nil {
				return nil, err
			}
			return core.Dial(ctx, inst, dest)
		},
	}

	c := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	return c, nil
}

// CoreHTTPRequest v2ray core http request
func CoreHTTPRequest(inst *core.Instance, timeout time.Duration, method, dest string) (int, []byte, error) {

	c, err := CoreHTTPClient(inst, timeout)
	if err != nil {
		return 0, nil, err
	}

	req, _ := http.NewRequest(method, dest, nil)
	resp, err := c.Do(req)
	if err != nil {
		return -1, nil, err
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, b, nil
}
