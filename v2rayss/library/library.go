//!build amd64
package main

import "C"

import (
	"encoding/json"
	"v2rayss"
)

var (
	app *v2rayss.App
)

//export newApp
func newApp() *C.char {
	var err error
	app, err = v2rayss.New()
	if err != nil {
		app = nil
		return C.CString("err: " + err.Error())
	}
	return C.CString("")
}

//export appRunning
func appRunning() C.uchar {
	if app == nil {
		return C.uchar(0)
	}
	if app.Running() {
		return C.uchar(1)
	}
	return C.uchar(0)
}

//export appConfInfo
func appConfInfo() *C.char {
	if app == nil {
		return C.CString(`{"addr":"", "port":0, "proto": "", "subaddr": ""}`)
	}
	d, err := json.Marshal(app.GetConfInfo())
	if err != nil {
		return C.CString("err: " +
			err.Error())
	}
	return C.CString(string(d))
}

//export appSetSubAddr
func appSetSubAddr(addr *C.char) *C.char {
	if app == nil {
		return C.CString("")
	}
	if err := app.SetSubAddr(C.GoString(addr)); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export appLoadSubAddr
func appLoadSubAddr() *C.char {
	if app == nil {
		return C.CString("")
	}
	if err := app.LoadSubAddr(); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export appPing
func appPing() *C.char {
	if app == nil {
		return C.CString("")
	}
	if err := app.PingLinks(); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export appListHosts
func appListHosts() *C.char {
	if app == nil {
		return C.CString("[]")
	}
	d, err := json.Marshal(app.ListHosts())
	if err != nil {
		return C.CString("err: " + err.Error())
	}
	return C.CString(string(d))
}

//export appSelectLink
func appSelectLink(index C.int) *C.char {
	if app == nil {
		return C.CString("[]")
	}
	if err := app.SelectLink(int(index)); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export appStart
func appStart() *C.char {
	if app == nil {
		return C.CString("")
	}
	if err := app.Start(); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export appClose
func appClose() *C.char {
	if app == nil {
		return C.CString("")
	}
	if err := app.Close(); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

func main() {}
