# build vars
TAG_NAME := $(shell test -d .git && git describe --abbrev=0 --tags)
VERSION := $(if $(TAG_NAME),$(TAG_NAME))
BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_TAGS := -tags=release
LD_FLAGS := -s -w
BUILD_ARGS := -trimpath -ldflags='$(LD_FLAGS)'

default:: build

all:: test build

install::
	go install $$(go list -e -f '{{join .Imports " "}}')

test::
	@echo "Running testsuite"
	CGO_ENABLED=0 go test $(BUILD_TAGS) ./...

build::
	@echo Version: $(VERSION) $(BUILD_DATE)
	CGO_ENABLED=0 go build -v $(BUILD_TAGS) $(BUILD_ARGS) 

upgrade::
	$(shell go list -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}{{end}}' -m all | xargs go get)
	go mod tidy
