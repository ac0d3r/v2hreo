package v2rayss

import (
	"fmt"
	"testing"
	"time"
)

func TestV2raySs(t *testing.T) {
	appt, err := New()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(appt.GetConfInfo())
	addr := "https://724.subadd.xyz/link/ct24m5Aea5M0oMIH?sub=3"
	t.Log(appt.SetSubAddr(addr))
	t.Log(appt.PingLinks())
	links := appt.ListHosts()
	for i := range links {
		fmt.Println(links[i])
	}
	fmt.Println("links - len:", len(links))

	t.Log(appt.SelectLink(10))

	t.Log("start......................")
	t.Log(appt.Start())

	time.Sleep(30 * time.Second)
	t.Log("change conf port:1081", appt.SaveConfInfo("127.0.0.1", "socks", 1081, "https://724.subadd.xyz/link/ct24m5Aea5M0oMIH?sub=3"))

	time.Sleep(30 * time.Second)
	t.Log("select an other one", appt.SelectLink(8))

	time.Sleep(30 * time.Second)
	t.Log(appt.Close())
}
