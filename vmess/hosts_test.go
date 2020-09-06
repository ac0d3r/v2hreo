package vmess

import "testing"

func TestParseSubscription(t *testing.T) {
	subURL := "https://sub/link/"
	t.Log(subURL)
	t.Log(ParseSubscription(subURL))
}
