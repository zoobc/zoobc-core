<h1 align="center">
  <a href="https://github.com/zoobc/zoobc-core">
    ZooBC Core
  </a>
</h1>
<p align="center">
  <a href="https://circleci.com/gh/zoobc/zoobc-core">
    <img src="https://circleci.com/gh/zoobc/zoobc-core.svg?style=svg&circle-token=cdd770bcb30a201696bb10e76ed15504cf235a9f" alt="CircleCI"/>
  </a>
</p>

### External Dependencies

- todo: specify external dependencies needed to run the code.

### Installation

- clone the repository.
- run `dep ensure -v --vendor-only` to install the dependencies.
- run `git submodule update --init --recursive --remote` to update submodule.

### Run

- run with `go run main.go`
- build `go build -o zoobc`

### Tests

- run all tests `go test ./...`
- run all test with coverage report `go test ./... -coverprofile=cover.out && go tool cover -html=cover.out`

### Swagger

- install `go-swagger` `https://github.com/go-swagger/go-swagger`
- pull newest `schema` and run `./compile-go.sh` to recompile the go file and produce swagger definition from it.
