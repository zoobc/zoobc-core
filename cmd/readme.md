### Zoobc - CMD

Command line interface to as a utility tools to develop the zoobc system.

### Structure

- main.go -> register all command
- package
    - specific.go
    - ...
- readme.md


### Run

- `go run main.go {command} {subcommand}`

- example: `go run main.go account generate` will generate account to use.