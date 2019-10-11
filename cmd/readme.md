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

### Genesis Generate From cmd/genesisblock/preRegisteredNodes.json

```
go run main.go genesis generate
outputs cmd/genesis.go.new and cmd/cluster_config.json
```

### Genesis Generate From cmd/genesisblock/preRegisteredNodes.json and resource/zoobc.db

```
go run main.go genesis generate -w
outputs cmd/genesis.go.new and cmd/cluster_config.json
```

### Genesis Generate From cmd/genesisblock/preRegisteredNodes.json and resource/zoobc.db, plus n random nodes registrations

```
go run main.go genesis generate -w -n 10
outputs cmd/genesis.go.new and cmd/cluster_config.json
```
