package vmess

import (
	"testing"
)

func TestPing(t *testing.T) {
	link := Link{
		Ps:   "",
		Host: "",
		Add:  "xx.x",
		Port: 724,
		ID:   "",
		Aid:  0,
		Net:  "ws",
		Type: "none",
		Path: "/v2ray",
		TLS:  "",
	}
	ts, err := link.Ping(1, "https://")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ts.String())
	t.Logf("%d", ts)
}
