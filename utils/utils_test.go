package utils

import "testing"

func TestGethome(t *testing.T) {
	t.Log(gethome())
}
func TestDirs(t *testing.T) {
	t.Log(TmpDir)
	t.Log(UserDir)
}

func TestCheckAppDir(t *testing.T) {
	t.Log(CheckAppDir(TmpDir, "runtime.log"))
	t.Log(CheckAppDir(UserDir, ".sub"))
}
