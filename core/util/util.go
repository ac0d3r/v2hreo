package util

import (
	"encoding/base64"
	"strings"
)

func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func Base64Decode(data string) (string, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decodeBytes), nil
}

var sep = map[rune]bool{
	' ':  true,
	'\n': true,
	',':  true,
	';':  true,
	'\t': true,
	'\f': true,
	'\v': true,
	'\r': true,
}

func StringSplit(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return sep[r]
	})
}
