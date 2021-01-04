package vmess

import (
	"testing"
)

func TestParseSubscription(t *testing.T) {
	subURL := "https://sub/link/"
	t.Log(subURL)
	t.Log(ParseSubscription(subURL))
}

func TestParseVmessURLList(t *testing.T) {
	sep := "\n"
	urls := "vmess://yaya=="
	t.Logf("%#v", parseVmessURLList(urls, sep))
}
