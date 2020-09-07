package utils

import (
	"os"
)

// CheckAppTmp check app tmp dir
func CheckAppTmp(filename string) (string, error) {
	tmpdir := "/tmp/buzz.V2raySs/"

	_, err := os.Stat(tmpdir)
	if os.IsNotExist(err) {
		err = os.Mkdir(tmpdir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	fn := tmpdir + filename
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
