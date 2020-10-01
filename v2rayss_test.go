package v2rayss

import "testing"

func TestPings(t *testing.T) {
	app = New()
	app.loadSubAddr()
	app.loadServerList()
	t.Logf("%v", app.serverList)
	app.Pings()
	t.Log(app.pings)
	t.Log(app.autoSelectServer())
}
