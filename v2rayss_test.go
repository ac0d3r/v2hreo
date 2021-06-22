package v2rayss

import (
	"testing"
	"time"
)

func TestV2raySs(t *testing.T) {
	appt, err := New()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(appt.GetConfInfo())
	addr := "https://xxxx"
	t.Log(appt.SetSubAddr(addr))
	t.Log(appt.PingLinks())
	t.Log(appt.ListHosts())
	t.Log(appt.SelectLink(10))
	t.Log(appt.Start())
	time.Sleep(50 * time.Second)
	t.Log("select an other one", appt.SelectLink(8))
	time.Sleep(50 * time.Second)
	t.Log(appt.Close())
}
