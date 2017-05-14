VERSION?=$(shell git describe --tags)

LD_FLAGS = -ldflags "-X github.com/tczekajlo/kir/utils.VERSION=$(VERSION) -s -w"

all: build

.PHONY: clean build
default: build
build: dist/kir

clean:
	rm -rf dist vendor

dist/kir:
	mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=linux go build $(LD_FLAGS) -v -o dist/kir

vendor:
	glide install --strip-vendor
