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

### Transaction Send Money

```
go run main.go generate transaction send-money --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --recipient VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM --amount 5000000000
```

### Transaction Register Node

```
go run main.go generate transaction register-node --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --node-owner-account-address VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM --node-seed "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness" --node-address "127.0.0.1" --locked-balance 1000000000
```

### Transaction Update Node Registration

```
go run main.go generate transaction update-node --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --node-owner-account-address VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM --node-seed "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness" --node-address "127.0.0.1" --locked-balance 10050000000000
```

### Transaction Claim Node

```
go run main.go generate transaction claim-node --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --node-owner-account-address VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM --node-seed "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness" --recipient VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM
```

### Transaction Remove Node

```
go run main.go generate transaction remove-node --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"  --node-seed "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness"
```

### Transaction Set Account Dataset

```
go run main.go generate transaction set-account-dataset --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --recipient VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM --property "Member" --value "Welcome to the jungle"
```

### Transaction Remove Account Dataset

```
go run main.go generate transaction remove-account-dataset --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --recipient VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM --property "Member" --value "Good Boy"
```
