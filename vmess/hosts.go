package vmess

import (
	"encoding/base64"
	"encoding/json"
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
	VerifyCert bool   `json:"verify_cert"`
	Address    string `json:"add"`
	Port       int    `json:"port"`
	Aid        int    `json:"aid"`
	Network    string `json:"net"`
	Type       string `json:"type"`
	Version    string `json:"v"`
	Ps         string `json:"ps"`
	ID         string `json:"id"`
	Class      int    `json:"class"`
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
//
func ParseSubscription(subURL string) ([]*Host, error) {
	resp, err := http.Get(subURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	urls, err := Base64Decode(string(body))
	if err != nil {
		return nil, err
	}
	return parseVmessURLList(urls, "\n"), nil
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
