// https://github.com/iochen/v2gen/blob/master/vmess/vmessping.go
package vmess

import (
	"encoding/json"
	"time"

	"v2rayss/core/util"

	"v2ray.com/core"
	"v2ray.com/core/app/dispatcher"
	applog "v2ray.com/core/app/log"
	"v2ray.com/core/app/proxyman"
	commlog "v2ray.com/core/common/log"
	"v2ray.com/core/common/serial"
)

var (
	NoPing time.Duration = -1
)

type Link struct {
	Ps   string      `json:"ps"`
	Host string      `json:"host"`
	Add  string      `json:"add"`
	Port interface{} `json:"port"`
	ID   string      `json:"id"`
	Aid  interface{} `json:"aid"`
	Net  string      `json:"net"`
	Type string      `json:"type"`
	Path string      `json:"path"`
	TLS  string      `json:"tls"`
}

func (lk *Link) Conf() string {
	if b, err := json.Marshal(lk); err != nil {
		return ""
	} else {
		return string(b)
	}
}

// String converts vmess link to vmess:// URL
func (lk *Link) String() string {
	return "vmess://" + util.Base64Encode(lk.Conf())
}

func (lk *Link) Ping(round int, dst string) (time.Duration, error) {
	server, err := startV2Ray(lk, false, true)
	if err != nil {
		return NoPing, err
	}
	defer server.Close()
	durationList := make([]time.Duration, 0, round)

	for count := 0; count < round; count++ {
		delay, err := measureDelay(server, 3*time.Second, dst)
		if err != nil {
			break
		}
		if delay > 0 {
			durationList = append(durationList, delay)
		}
	}

	if len(durationList) == 0 {
		return NoPing, nil
	}
	//take the average
	return delayAverage(durationList), nil
}

func delayAverage(list []time.Duration) time.Duration {
	delay := time.Duration(0)
	for _, d := range list {
		delay += d
	}
	u := uint64(time.Duration(int(delay) / len(list)))
	if u > uint64(time.Second) {
		u = u / 10000000 * 10000000
	} else {
		u = u / 100000 * 100000
	}
	return time.Duration(u)
}

func startV2Ray(lk *Link, verbose, usemux bool) (*core.Instance, error) {
	loglevel := commlog.Severity_Error
	if verbose {
		loglevel = commlog.Severity_Debug
	}

	out, err := Vmess2Outbound(lk, usemux)
	if err != nil {
		return nil, err
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
		Outbound: []*core.OutboundHandlerConfig{out},
	}
	server, err := core.New(config)
	if err != nil {
		return nil, err
	}

	return server, nil
}
