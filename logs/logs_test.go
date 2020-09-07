package logs

import (
	"testing"
)

func TestLogger(t *testing.T) {
	defer Close()
	Info("Printt")
	Fatalln("Fatal")
}
