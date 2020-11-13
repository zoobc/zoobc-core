BIN_DIR := $(GOPATH)/bin
GOLANGCILINT := $(BIN_DIR)/golangci-lint
GOLANGCILINT_VERSION := v1.23.8
XGO := $(BIN_DIR)/xgo
VERSION ?= latest
ZBCPATH ?= dist
BINARY_CORE_NAME := zoobc
BINARY_CLI_NAME := zcmd
CORE_OUPUT := $(ZBCPATH)/$(BINARY_CORE_NAME)-$(VERSION)
CLI_OUPUT := $(ZBCPATH)/$(BINARY_CLI_NAME)-$(VERSION)
GITHUB_TOKEN ?= $(shell cat github.token)
genesis := false
gen-target:= alpha
gen-output := resource

.PHONY: test
test: go-fmt golangci-lint
	$(info    running unit tests...)
	go test `go list ./... | egrep -v 'common/model|common/service'` --short -count=1

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
	mkdir -p $(ZBCPATH)
	go build -o release/$(BINARY_CORE_NAME)-$(VERSION)

.PHONY: core-linux
core-linux: $(XGO)
	$(info    build core with linux as target...)
	mkdir -p $(ZBCPATH)
	xgo --targets=linux/amd64 -out=$(CORE_OUPUT) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./

.PHONY: core-arm
core-arm: $(XGO)
	$(info    build core with linux/arm as target...)
	mkdir -p $(ZBCPATH)
	xgo --targets=linux/arm -out=$(CORE_OUPUT) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./

.PHONY: core-windows
core-windows: $(XGO)
	$(info    build core with windows as target...)
	mkdir -p $(ZBCPATH)
	xgo --targets=windows/* -out=$(CORE_OUPUT) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./

.PHONY: core-darwin
core-darwin: $(XGO)
	$(info    build core with darwin/macos as target...)
	mkdir -p $(ZBCPATH)
	xgo --targets=darwin/* -out=$(CORE_OUPUT) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./

.PHONY: cmd-darwin
cmd-darwin: $(XGO)
	$(info    build cmd with darwin/macos as target...)
	mkdir -p $(ZBCPATH)
	xgo --targets=darwin/* -out=$(CLI_OUPUT) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./cmd/

.PHONY: cmd-linux
cmd-linux: $(XGO)
	$(info    build cmd with linux as target...)
	mkdir -p $(ZBCPATH)
	xgo --targets=linux/amd64 -out=$(CLI_OUPUT) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./cmd/

.PHONY: cmd-arm
cmd-arm: $(XGO)
	$(info    build cmd with linux/arm as target...)
	mkdir -p $(ZBCPATH)
	xgo --targets=linux/arm -out=$(CLI_OUPUT) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./cmd/

.PHONY: cmd-windows
cmd-windows: $(XGO)
	$(info    build cmd with windows as target...)
	mkdir -p $(ZBCPATH)
	xgo --targets=windows/* -out=$(CLI_OUPUT) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./cmd/


.PHONY: core-common-os
core-common-os: $(XGO)
	$(info    build core with windows, linux & darwin as targets...)
	mkdir -p $(ZBCPATH)/linux $(ZBCPATH)/windows $(ZBCPATH)/darwin
	xgo --targets=windows/*,linux/amd64,darwin/amd64 -out=$(CORE_OUPUT) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./
	mv $(CORE_OUPUT)-linux* $(ZBCPATH)/linux/$(BINARY_CORE_NAME)
	mv $(CORE_OUPUT)-windows*386.exe $(ZBCPATH)/windows/$(BINARY_CORE_NAME)-32bit.exe
	mv $(CORE_OUPUT)-windows*amd64.exe $(ZBCPATH)/windows/$(BINARY_CORE_NAME)-64bit.exe
	mv $(CORE_OUPUT)-darwin* $(ZBCPATH)/darwin/$(BINARY_CORE_NAME)

.PHONY: cmd-common-os
cmd-common-os: $(XGO)
	$(info    build cmd with windows, linux & darwin as targets...)
	mkdir -p $(ZBCPATH)/linux $(ZBCPATH)/windows $(ZBCPATH)/darwin
	xgo --targets=windows/*,linux/amd64,darwin/amd64 -out=$(CLI_OUPUT) --go-private=github.com/zoobc/* --github-token=$(GITHUB_TOKEN)  ./cmd/
	mv $(CLI_OUPUT)-linux* $(ZBCPATH)/linux/$(BINARY_CLI_NAME)
	mv $(CLI_OUPUT)-windows*386.exe $(ZBCPATH)/windows/$(BINARY_CLI_NAME)-32bit.exe
	mv $(CLI_OUPUT)-windows*amd64.exe $(ZBCPATH)/windows/$(BINARY_CLI_NAME)-64bit.exe
	mv $(CLI_OUPUT)-darwin* $(ZBCPATH)/darwin/$(BINARY_CLI_NAME)

.PHONY: build-all-common
build-all-common: core-common-os cmd-common-os

.PHONY: release-core
release-core: core-linux

.PHONY: release-cmd
release-cmd: cmd-linux

.PHONY: reset-data
reset-data: 
	rm -rf resource/*db resource/snapshots*
