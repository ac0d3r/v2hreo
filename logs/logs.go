package logs

import (
	oslog "log"
	"os"
	"sync"
	"v2rayss/utils"
)

var (
	out    *os.File
	once   sync.Once
	logger *oslog.Logger
)

func init() {
	logger = NewLogger(false)
}

// NewLogger return an instance of v2rayss app
func NewLogger(debug bool) *oslog.Logger {
	once.Do(func() {
		if debug == true {
			logger = oslog.New(os.Stderr, "", oslog.LstdFlags)
		} else {
			fn, err := utils.CheckAppTmp("runtime.log")
			out, err := os.Create(fn)
			if err != nil {
				logger = oslog.New(os.Stderr, "", oslog.LstdFlags)
			}
			logger = oslog.New(out, "", oslog.LstdFlags)
		}

	})
	return logger
}

// Close logger out file
func Close() {
	if out != nil {
		out.Close()
	}
}

// Info logger info
func Info(v ...interface{}) {
	logger.Println(v...)
}

// Fatalln logger info
func Fatalln(v ...interface{}) {
	logger.Fatalln(v...)
}
