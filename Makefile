.PHONY: all install clean

PREFIX ?= /usr/local
GOFILES := $(shell find . -type f -name \*.go)
RESOURCES := $(shell find static/resources -type f)

all: qemuer

qemuer: static/blob.go $(GOFILES)
	cd cmd/qemuer; go generate
	go build ./cmd/qemuer

static/blob.go: static/gen/gen.go $(RESOURCES)
	cd static; go generate

install: qemuer
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 755 qemuer $(DESTDIR)$(PREFIX)/bin

clean:
	rm -fv qemuer static/blob.go cmd/qemuer/version-number.go
