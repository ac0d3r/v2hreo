all: build/macos

build/macos:
	cd v2rayss && CGO_ENABLED=1 go build --buildmode=c-archive  -o libv2rayss.a library/library.go
	mv v2rayss/libv2rayss.h v2rayss/libv2rayss.a ui/v2hreo/
clean:
	rm ui/libv2rayss.*
