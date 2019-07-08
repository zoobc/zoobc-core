## ZooBC-Core

### External Dependencies

- todo: specify external dependencies needed to run the code.

### Installation

- clone the repository.
- run `dep ensure -v --vendor-only` to install the dependencies.
- run `git submodule update --init --recursive --remote` to update submodule.

### Run

- run with `go run main.go`
- build `go build main.go`

### Tests

- run all tests `go test ./...`
- run all test with coverage report `go test ./... -coverprofile=cover.out && go tool cover -html=cover.out`

### Swagger

- install `go-swagger` `https://github.com/go-swagger/go-swagger`
- pull newest `schema` and run `./compile-go.sh` to recompile the go file and produce swagger definition from it.
