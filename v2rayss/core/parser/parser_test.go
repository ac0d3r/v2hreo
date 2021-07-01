package parser

import "testing"

func TestParse(t *testing.T) {
	links, err := Parse("https://xxx")
	if err != nil {
		t.Fatal(err)
	}

	for _, l := range links {
		t.Logf("%#v", l)
		t.Log(l.Ping(3, "https://cloudflare.com/cdn-cgi/trace"))
	}
}
