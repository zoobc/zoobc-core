<h1 align="center">
  <a href="https://github.com/zoobc/zoobc-core">
    ZOOBC Core
  </a>
</h1>
<p align="center">
  <a href="https://circleci.com/gh/zoobc/zoobc-core">
    <img src="https://circleci.com/gh/zoobc/zoobc-core.svg?style=svg&circle-token=cdd770bcb30a201696bb10e76ed15504cf235a9f" alt="CircleCI"/>
  </a>
  <a href="#">
    <img src="./coverage_badge.png" alt="cover badge"/>
  </a>
</p>

> Zoobc-core is the main node application to run the zoobc blockchain. This repository consist of the main node application and the `command line interface` tools to help with development, which is located in the `cmd/` directory.

Table of Contents:

-   [Environments](#environments)
-   [Install](#install)
-   [Build](#build)
-   [Run](#run)
-   [Tests](#tests)
-   [Swagger](#swagger)
-   [Contributing](#contributing)
-   [GRPC web proxy for browser](#grpc-web-proxy-for-browser)

### Environments

-   [golang](https://golang.org/doc/install), currently using go.1.14
-   [go-swagger](https://github.com/go-swagger/go-swagger) optional. Used as tools to document the rpc endpoint.
-   [gopherbadger](https://github.com/jpoles1/gopherbadger) optional. Used to calculate total test coverage.
-   [protoc](https://github.com/protocolbuffers/protobuf), optional as we are pushing the generated go file to the repo.
-   [protoc-gen-go](https://github.com/golang/protobuf), optional as we are pushing the generated go file to the repo.
-   [golangci-lint](https://github.com/golangci/golangci-lint) lint tools we used to keep the code clean and well structured.

### Install

-   clone the repository.
-   Dep user: run `dep ensure -v --vendor-only` to install the dependencies read from Gopkg.toml only.
-   Go mod user: `go mod download` to generate vendor directory which is should download the packages from `Gopkg.toml` or read from project recursively
-   run `git submodule update --init --recursive --remote` to update / fetch submodule.
-   run `make test` to run the test and linter.
    VSCode go modules support with this config:

```json
"go.useLanguageServer": true
```

### Build

To make use of the `Makefile` please rename `github.token.example` to `github.token` and place your github token there. This is required since we are accessing private repository for one of our dependencies.

-   Core

    note: For cross compilation please install and activate docker.

    For:

    -   host: `go build -o zoobc`
    -   darwin: `make VERSION=v1.10.1 core-darwin`
    -   linux (386 & amd64): `make VERSION=v1.10.1 core-linux`
    -   windows (32 & 64bit): `make VERSION=v1.10.1 core-windows`

-   CMD

    For:

    -   host: `go build -o zoobc`
    -   darwin: `make VERSION=v1.10.1 cmd-darwin`
    -   linux (386 & amd64): `make VERSION=v1.10.1 cmd-linux`
    -   windows (32 & 64bit): `make VERSION=v1.10.1 cmd-windows`

### Run

> If already build, just run the binary

```bash
./zoobc
```

> Main node application run manually

```bash
go run main.go
```

-   Flags:
    1. `debug` (bool): enters debug mode with capabilities like `prometheus monitoring` (uses port defined by `monitoringPort` in the config file).
    2. `config-postfix` (string): defines which config file with defined postfix to use. It will use the config file in `./resource/config{postfix}`.toml.
    3. `config-path` (string): defines the directory that will hold the resources and configs. by default it will use `./resource`. This will be useful particularly for robust resource placements, for example while used in Zoobc-Testing-Framework.
    4. `cpu-profile` (bool): enable realtime profiling for the running instance, via http server.
    See [http pprof](https://golang.org/pkg/net/http/pprof/) for documentation on how to use this tool 
    
-   Command line tools
    ```bash
    cd cmd
    go run main.go --help
    go run main.go tx generate -t registerNode
    ```
    for more detail check this out [cmd session](https://github.com/zoobc/zoobc-core/tree/develop/cmd)

### Tests

-   #### Unit tests
    -   run all tests without cache `go test ./... -count=1`
    -   run all test with coverage report `go test ./... -coverprofile=cover.out && go tool cover -html=cover.out`
-   #### Linter
    -   install linter tools `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.20.0`
    -   run `golangci-lint run` to check any linting error in the changes.
    -   remember to run tests, and lint before submitting PR.

### Swagger

-   install
-   pull newest `schema` and run `./compile-go.sh` to recompile the go file and produce swagger definition from it.

### Contributing

please refer to [contribute.md](contribute.md) and [code of conduct](code_of_conduct.md).

### GRPC web proxy for browser

[GRPC Web Proxy](https://github.com/improbable-eng/grpc-web/tree/master/go/grpcwebproxy)

```
nohup grpcwebproxy --backend_addr=localhost:7000 --run_tls_server=false --allow_all_origins --server_http_debug_port=7001 --server_http_max_write_timeout 1h &
```
