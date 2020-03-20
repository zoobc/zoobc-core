BIN_DIR := $(GOPATH)/bin
GOLANGCILINT := $(BIN_DIR)/golangci-lint
GOLANGCILINT_VERSION := v1.20.0
XGO := $(BIN_DIR)/xgo
VERSION ?= latest
BINARY_CORE := zoobc
BINARY_CLI := zoomd
GITHUB_TOKEN ?= $(shell cat github.token)

.PHONY: test
test: go-fmt golangci-lint
	go test `go list ./... | egrep -v 'common/model|common/service'` --short

$(GOLANGCILINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin $(GOLANGCILINT_VERSION)

$(XGO):
	go get github.com/zoobc/xgo

.PHONY: golangci-lint
golangci-lint: $(GOLANGCILINT)
	golangci-lint run

.PHONY: go-fmt
go-fmt:
	go fmt `go list ./... | egrep -v 'common/model|common/service|vendor'`

.PHONY: build
build:
	mkdir -p release
	go build -o release/$(BINARY_CORE)-$(VERSION)

.PHONY: core-linux
core-linux: $(XGO)
	mkdir -p release
	xgo --targets=linux/amd64 -out=release/$(BINARY_CORE)-$(VERSION) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./

.PHONY: core-windows
core-windows: $(XGO)
	mkdir -p release
	xgo --targets=windows/* -out=release/$(BINARY_CORE)-$(VERSION) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./

.PHONY: core-darwin
core-darwin: $(XGO)
	mkdir -p release
	xgo --targets=darwin/* -out=release/$(BINARY_CORE)-$(VERSION) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./

.PHONY: cmd-darwin
cmd-darwin: $(XGO)
	mkdir -p cmd/release
	xgo --targets=darwin/* -out=cmd/release/$(BINARY_CORE)-$(VERSION) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./cmd/

.PHONY: cmd-linux
cmd-linux: $(XGO)
	mkdir -p release
	xgo --targets=linux/amd64 -out=cmd/release/$(BINARY_CORE)-$(VERSION) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./

.PHONY: cmd-windows
cmd-windows: $(XGO)
	mkdir -p release
	xgo --targets=windows/* -out=cmd/release/$(BINARY_CORE)-$(VERSION) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./

.PHONY: release-core
release-core: core-linux

.PHONY: release-cmd
release-cmd: cmd-linux
