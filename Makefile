BIN_DIR := $(GOPATH)/bin
GOLANGCILINT := $(BIN_DIR)/golangci-lint
GOLANGCILINT_VERSION := v1.23.8
XGO := $(BIN_DIR)/xgo
VERSION ?= latest
BINARY_CORE := zoobc
BINARY_CLI := zoomd
GITHUB_TOKEN ?= $(shell cat github.token)
genesis := false
gen-target:= alpha
gen-output := resource

.PHONY: test
test: go-fmt golangci-lint
	$(info    running unit tests...)
	go test `go list ./... | egrep -v 'common/model|common/service'` --short

$(GOLANGCILINT):
	$(info    fetching golangci-lint...)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin $(GOLANGCILINT_VERSION)

$(XGO):
	$(info    fetching zoobc/xgo...)
	go get github.com/zoobc/xgo

.PHONY: golangci-lint
golangci-lint: $(GOLANGCILINT)
	$(info    running linter...)
	golangci-lint run --timeout=20m -v

.PHONY: go-fmt
go-fmt:
	$(info    running go-fmt...)
	go fmt `go list ./... | egrep -v 'common/model|common/service|vendor'`

.PHONY: generate-gen
generate-gen:
	$(info generating new genesis file and replace old genesis file ...)
	go run cmd/main.go genesis generate -e ${gen-target} -o ${gen-output}
	cp ./${gen-output}/generated/genesis/genesis*.go ./common/constant/


.PHONY: build
build:
ifdef genesis
	$(MAKE) generate-gen
endif
	$(info    build core with host os as target...)
	mkdir -p release
	go build -o release/$(BINARY_CORE)-$(VERSION)

.PHONY: core-linux
core-linux: $(XGO)
	$(info    build core with linux as target...)
	mkdir -p release
	xgo --targets=linux/amd64 -out=release/$(BINARY_CORE)-$(VERSION) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./

.PHONY: core-windows
core-windows: $(XGO)
	$(info    build core with windows as target...)
	mkdir -p release
	xgo --targets=windows/* -out=release/$(BINARY_CORE)-$(VERSION) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./

.PHONY: core-darwin
core-darwin: $(XGO)
	$(info    build core with darwin/macos as target...)
	mkdir -p release
	xgo --targets=darwin/* -out=release/$(BINARY_CORE)-$(VERSION) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./

.PHONY: cmd-darwin
cmd-darwin: $(XGO)
	$(info    build cmd with darwin/macos as target...)
	mkdir -p cmd/release
	xgo --targets=darwin/* -out=cmd/release/$(BINARY_CORE)-$(VERSION) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./cmd/

.PHONY: cmd-linux
cmd-linux: $(XGO)
	$(info    build cmd with linux as target...)
	mkdir -p release
	xgo --targets=linux/amd64 -out=cmd/release/$(BINARY_CORE)-$(VERSION) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./

.PHONY: cmd-windows
cmd-windows: $(XGO)
	$(info    build cmd with windows as target...)
	mkdir -p release
	xgo --targets=windows/* -out=cmd/release/$(BINARY_CORE)-$(VERSION) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./

.PHONY: release-core
release-core: core-linux

.PHONY: release-cmd
release-cmd: cmd-linux
