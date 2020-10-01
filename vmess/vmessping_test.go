package vmess

import (
	"context"
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
	round := 3
	//test destination url (vmess ping only)
	dst := "https://cloudflare.com/cdn-cgi/trace"
	info := `{"host":"123.123.12.1","path":"/","tls":"","verify_cert":true,"add":"address","port":443,"aid":0,"net":"ws","type":"none","v":"1","ps":"some infos","id":"server-id","class":1}`
	host := Host{}
	err := json.Unmarshal([]byte(info), &host)

	ctx, cannel := context.WithTimeout(context.Background(), 3*time.Duration(round)*time.Second)
	defer cannel()
	delay, err := Ping(ctx, &host, round, dst)
	t.Log("Ping timeout:", delay, "error:", err)
}
