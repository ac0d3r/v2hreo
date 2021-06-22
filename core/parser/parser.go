package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"v2rayss/core/util"
	"v2rayss/core/vmess"
)

const (
	vmessProtocol = "vmess://"
)

var (
	ErrWrongProtocol = errors.New("wrong protocol")
)

func Parse(addr string) ([]*vmess.Link, error) {
	var raw string = addr
	// 处理 http(s)?://xxx 的订阅地址
	if strings.HasPrefix(addr, "http") {
		var (
			resp *http.Response
			body []byte
			err  error
		)
		if resp, err = http.Get(addr); err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if body, err = ioutil.ReadAll(resp.Body); err != nil {
			return nil, err
		}
		if raw, err = util.Base64Decode(string(body)); err != nil {
			return nil, err
		}
	}
	return ParseVmessURLs(raw), nil
}

func ParseVmessURLs(s string) []*vmess.Link {
	hosts := make([]*vmess.Link, 0, 4)
	for _, vmessURL := range util.StringSplit(s) {
		h, err := ParseVmessURL(vmessURL)
		if err == nil {
			hosts = append(hosts, h)
		}
	}
	return hosts
}

// ParseVmessURL 处理 vmess://{base64(info)}
func ParseVmessURL(vmessURL string) (*vmess.Link, error) {
	if len(vmessURL) < len(vmessProtocol) {
		return nil, fmt.Errorf("wrong url:%s", vmessURL)
	}
	if !strings.HasPrefix(vmessURL, vmessProtocol) {
		return nil, ErrWrongProtocol
	}

	raw, err := util.Base64Decode(vmessURL[len(vmessProtocol):])
	if err != nil {
		return nil, err
	}

	h := &vmess.Link{}
	err = json.Unmarshal([]byte(raw), h)
	return h, err
}
