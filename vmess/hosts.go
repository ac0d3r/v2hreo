package vmess

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	// VmessProtocol vmess protocol
	VmessProtocol = "vmess://"
)

// Host provided by subscription address
type Host struct {
	Host       string `json:"host"`
	Path       string `json:"path"`
	TLS        string `json:"tls"`
	VerifyCert bool   `json:"verify_cert,omitempty"`
	Address    string `json:"add"`
	Port       int    `json:"port"`
	Aid        int    `json:"aid"`
	Network    string `json:"net"`
	Type       string `json:"type"`
	Version    int    `json:"v"`
	Ps         string `json:"ps"`
	ID         string `json:"id"`
	Class      int    `json:"class,omitempty"`
}

// Base64Decode decode base64
func Base64Decode(data string) (string, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decodeBytes), nil
}

// ParseSubscription parseing v2ray subscription address
func ParseSubscription(subURL string) ([]*Host, error) {
	var (
		vmessURLList string
		body         []byte
		resp         *http.Response
		err          error
	)
	// http 协议
	if strings.HasPrefix(subURL, "http") {
		if resp, err = http.Get(subURL); err != nil {
			return nil, err
		}
		if body, err = ioutil.ReadAll(resp.Body); err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if vmessURLList, err = Base64Decode(string(body)); err != nil {
			return nil, err
		}
		return parseVmessURLList(vmessURLList, "\n"), nil
	} else if strings.HasPrefix(subURL, VmessProtocol) {
		return parseVmessURLList(subURL, "\n"), nil
	}
	return nil, errors.New("Protocol not supported")
}

func parseVmessURLList(vmessURLList, sep string) []*Host {
	hosts := []*Host{}
	for _, vmessURL := range strings.Split(vmessURLList, sep) {
		// parse single vmess url
		if len(vmessURL) < len(VmessProtocol) || !strings.HasPrefix(vmessURL, VmessProtocol) {
			continue
		}
		v, _ := Base64Decode(vmessURL[len(VmessProtocol):])
		host := Host{}
		err := json.Unmarshal([]byte(v), &host)
		if err == nil {
			hosts = append(hosts, &host)
		}
	}
	return hosts
}
