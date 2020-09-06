package vmess

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDelayAverage(t *testing.T) {
	list := []time.Duration{time.Duration(12440), time.Duration(14442), time.Duration(123444)}
	t.Log("Average:", delayAverage(list))
}

func TestPing(t *testing.T) {
	// ping count for each node
	c := 3
	//test destination url (vmess ping only)
	dst := "https://cloudflare.com/cdn-cgi/trace"
	info := `{"host":"","path":"/","tls":"","verify_cert":true,"add":"address","port":443,"aid":0,"net":"ws","type":"none","v":"1","ps":"some infos","id":"server-id","class":1}`
	host := Host{}
	err := json.Unmarshal([]byte(info), &host)

	delay, err := Ping(&host, c, dst)
	t.Log("Ping timeout:", delay, "error:", err)
}
