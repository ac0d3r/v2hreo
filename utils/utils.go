package utils

import (
	"os"
	"os/user"
	"path"
	"strings"
)

var (
	appname = "V2raySs"
	// TmpDir tmp dir
	TmpDir = getpath("/tmp", appname)
	// UserDir user dir
	UserDir = getpath(gethome(), appname)
)

// CheckAppDir check app tmp dir
func CheckAppDir(dir, filename string) (string, error) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	fn := dir + filename
	_, err = os.Stat(fn)
	if os.IsNotExist(err) {
		f, err := os.Create(fn)
		defer f.Close()
		if err != nil {
			return "", err
		}
	}
	return fn, nil
}

func gethome() string {
	user, err := user.Current()
	if err == nil {
		return user.HomeDir
	}
	return ""
}

func getpath(base, appname string) string {
	if !strings.HasPrefix(appname, ".") {
		appname = "." + appname
	}
	return path.Join(base, appname)
}
