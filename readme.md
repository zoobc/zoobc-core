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

**Table of Contents**:

- [Environments](#environments)
- [Development](#development)
  - [Build](#build)
  - [Run](#run)
  - [Flags:](#flags)
- [Tests](#tests)
- [Installation script](#installation-script)
- [Swagger](#swagger)
- [Contributing](#contributing)

### Environments

- [golang](https://golang.org/doc/install), currently using go.1.14
- [go-swagger](https://github.com/go-swagger/go-swagger) optional. Used as tools to document the rpc endpoint.
- [protoc](https://github.com/protocolbuffers/protobuf), optional as we are pushing the generated go file to the repo. Version compatible: [this](https://github.com/protocolbuffers/protobuf/releases/tag/v3.12.4)
- [protoc-gen-go](https://github.com/golang/protobuf), optional as we are pushing the generated go file to the repo. Instruction: [here](https://grpc.io/docs/languages/go/quickstart/)
- [golangci-lint](https://github.com/golangci/golangci-lint) lint tools we used to keep the code clean and well structured.
- [tdm-gcc](https://jmeubank.github.io/tdm-gcc/) (windows-only), if you are running on windows you'll need to install this to build the binary.
- [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway/tree/v1) (optional), for reverse proxy server.

### Development

- clone the repository.
- Dep user: run `dep ensure -v --vendor-only` to install the dependencies read from Gopkg.toml only.
- Go mod user: `go mod download` to generate vendor directory which is should download the packages from `Gopkg.toml` or read from project recursively
- run `git submodule update --init --recursive --remote` to update / fetch submodule.
- run `make test` to run the test and linter.

#### Build

To make use of the `Makefile` please rename `github.token.example` to `github.token` and place your github token there. This is required since we are accessing private repository for one of our dependencies.

- ZOOBC CORE

  > note: For cross compilation please install and activate docker.
  > For:

  - host: `go build -o zoobc`
  - darwin: `make VERSION=v1.10.1 core-darwin`
  - linux (386 & amd64): `make VERSION=v1.10.1 core-linux`
  - windows (32 & 64bit): `make VERSION=v1.10.1 core-windows`
  - common os (darwin, linux, windows) : `make VERSION=v1.10.1 core-common-os`
    > With genesis replacement, you can add argument `genesis=true` and what your target is {develop,staging,alhpa(default),local}, like:
    > `make build genesis=true gen-target=develop gen-output=dist` for the local target you need create `local.preRegisteredNodes.json`.

- ZOOBC CMD

  For:

  - host: `go build -o zoobc`
  - darwin: `make VERSION=v1.10.1 cmd-darwin`
  - linux (386 & amd64): `make VERSION=v1.10.1 cmd-linux`
  - windows (32 & 64bit): `make VERSION=v1.10.1 cmd-windows`
  - common os (darwin, linux, windows) : `make VERSION=v1.10.1 cmd-common-os`

#### Run

```bash
Usage:
   [command]

Available Commands:
  daemon      Run node on daemon service, which mean running in the background. Similar to launchd or systemd
  help        Help about any command
  run         Run node as without daemon.

Flags:
      --config-path string      Configuration path (default "./")
      --config-postfix string   Configuration version
      --debug                   Run on debug mode
  -h, --help                    help for this command
      --profiling               Run with profiling
      --use-env                 Running node without configuration file

Use " [command] --help" for more information about a command.
```

#### Flags:

- `debug` (bool): enters debug mode with capabilities like `prometheus monitoring` (uses port defined by `monitoringPort` in the config file).
- `config-postfix` (string): defines which config file with defined postfix to use. It will use the config file in `./resource/config{postfix}`.toml.
- `config-path` (string): defines the directory that will hold the resources and configs. by default it will use `./resource`. This will be useful particularly for robust resource placements, for example while used in Zoobc-Testing-Framework.
- `cpu-profile` (bool): enable realtime profiling for the running instance, via http server.
  See [http pprof](https://golang.org/pkg/net/http/pprof/) for documentation on how to use this tool
- Command line tools
  ```bash
  cd cmd
  go run main.go --help
  go run main.go tx generate -t registerNode
  ```
  for more detail check this out [cmd session](https://github.com/zoobc/zoobc-core/tree/develop/cmd)

### Tests

- #### Unit tests
  - run all tests without cache `go test ./... -count=1`
  - run all test with coverage report `go test ./... -coverprofile=cover.out && go tool cover -html=cover.out`
- #### Linter
  - install linter tools `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.20.0`
  - run `golangci-lint run` to check any linting error in the changes.
  - remember to run tests, and lint before submitting PR.

### Installation script

ZooBC installation script based on **_bashscript_** for help user to install ZooBC node in one easy way. It will download the latest binary that contains:

- `zcmd`: command line tools helper for developer and that needed for check and parse wallet certificate and stuff at first time running the installation script.
- `zoobc`: zoobc core binary
  Both will downloaded by the script and for more detail you can check [here](https://github.com/zoobc/zoobc-installer).

### Swagger

- install
- pull newest `schema` and run `./compile-go.sh` to recompile the go file and produce swagger definition from it.

### Contributing

please refer to [contribute.md](contribute.md) and [code of conduct](code_of_conduct.md).
