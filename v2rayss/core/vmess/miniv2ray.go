package vmess

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"v2ray.com/core"
	v2net "v2ray.com/core/common/net"
	"v2ray.com/core/infra/conf"
)

func Vmess2Inbound(addr, protocol string, port uint32) (*core.InboundHandlerConfig, error) {
	in := &conf.InboundDetourConfig{}
	v := conf.Address{}
	v.Address = v2net.ParseAddress(addr)
	in.Tag = "proxy"
	in.ListenOn = &v
	in.Protocol = protocol
	in.PortRange = &conf.PortRange{From: port, To: port}
	inset := json.RawMessage([]byte(`{"udp": true}`))
	in.Settings = &inset
	return in.Build()
}

func Vmess2Outbound(link *Link, usemux bool) (*core.OutboundHandlerConfig, error) {
	out := &conf.OutboundDetourConfig{}
	out.Tag = "proxy"
	out.Protocol = "vmess"
	out.MuxSettings = &conf.MuxConfig{}
	if usemux {
		out.MuxSettings.Enabled = true
		out.MuxSettings.Concurrency = 8
	}

	p := conf.TransportProtocol(link.Net)
	s := &conf.StreamConfig{
		Network:  &p,
		Security: link.TLS,
	}

	switch link.Net {
	case "tcp":
		s.TCPSettings = &conf.TCPConfig{}
		if link.Type == "" || link.Type == "none" {
			s.TCPSettings.HeaderConfig = json.RawMessage([]byte(`{ "type": "none" }`))
		} else {
			pathb, _ := json.Marshal(strings.Split(link.Path, ","))
			hostb, _ := json.Marshal(strings.Split(link.Host, ","))
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
		s.KCPSettings.HeaderConfig = json.RawMessage([]byte(fmt.Sprintf(`{ "type": "%s" }`, link.Type)))
	case "ws":
		s.WSSettings = &conf.WebSocketConfig{}
		s.WSSettings.Path = link.Path
		s.WSSettings.Headers = map[string]string{
			"Host": link.Host,
		}
	case "h2", "http":
		s.HTTPSettings = &conf.HTTPConfig{
			Path: link.Path,
		}
		if link.Host != "" {
			h := conf.StringList(strings.Split(link.Host, ","))
			s.HTTPSettings.Host = &h
		}
	}

	if link.TLS == "tls" {
		s.TLSSettings = &conf.TLSConfig{
			Insecure: true,
		}
		if link.Host != "" {
			s.TLSSettings.ServerName = link.Host
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
}`, link.Add, link.Port, link.ID, link.Aid)))
	out.Settings = &oset
	return out.Build()
}

func measureDelay(inst *core.Instance, timeout time.Duration, dest string) (time.Duration, error) {
	start := time.Now()
	code, err := CoreHTTPRequest(inst, timeout, "GET", dest)
	if err != nil {
		return -1, err
	}
	if code >= 4e2 {
		return -1, fmt.Errorf("status incorrect (>= 400): %d", code)
	}
	return time.Since(start), nil
}

func CoreHTTPRequest(inst *core.Instance, timeout time.Duration, method, dest string) (int, error) {
	if inst == nil {
		return 0, errors.New("core instanc")
	}
	http.DefaultClient.Timeout = timeout
	tr := http.DefaultTransport.(*http.Transport)

	keepAlives := tr.DisableKeepAlives
	dialCtx := tr.DialContext
	defer func() {
		tr.DisableKeepAlives = keepAlives
		tr.DialContext = dialCtx
		http.DefaultClient.Timeout = 0
	}()
	tr.DisableKeepAlives = true
	tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		dest, err := v2net.ParseDestination(fmt.Sprintf("%s:%s", network, addr))
		if err != nil {
			return nil, err
		}
		return core.Dial(ctx, inst, dest)
	}

	req, err := http.NewRequest(method, dest, nil)
	if err != nil {
		return -1, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}
