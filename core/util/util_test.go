package util

import "testing"

func TestBase64Encode(t *testing.T) {
	t.Log(Base64Encode("123123"))
}
func TestBase64Decode(t *testing.T) {
	t.Log(Base64Decode("MTIzMTIz"))
}

func TestStringSplit(t *testing.T) {
	t.Log(StringSplit("sdasd asds\n asdasd\t asdasdasd,dasdasd;"))
}
