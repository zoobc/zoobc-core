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

### Transaction general flag 
  - `--output` to provide generated ouput type. Example: `--ouput bytes`
  - `--version` to provide version of transaction. Example: `--version 1`
  - `--timestamp` to provide timestamp of trasaction. Example: `--timestamp 1234567`
  - `--sender-seed` to provide the seed of sender transaction. Example: `--sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"`
  - `--recipient` provide recepient transaction. Example `--recipient VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM`
  - `--fee` to provide fee transaction, Example `--fee 1`
  - `--post` to define automate post transaction or not. Example: `-post true`
  - `--post-host` to provide where the transaction will post. Example: `--post-host "127.0.0.1:7000"`
  - `--sender-signature-type` to provide type of transaction signature and effected to the type of the sender account. Example: `--sender-signature-type 1`


### Transaction Send Money

```
go run main.go generate transaction send-money --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --recipient VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM --amount 5000000000
```
#### Transaction send money escrow, set flag `--escrow true` and 3 more fields: `--approver-address`, `--timeout`, `--commission` and `--instruction`
```bash
go run main.go generate transaction send-money --escrow true --approver-address BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE --timeout 200 --sender-seed "execute beach inflict session course dance vanish cover lawsuit earth casino fringe waste warfare also habit skull donate window cannon scene salute dawn good" --amount 1111 --commission 111 --instruction "Check amount should be 111" --recipient nK_ouxdDDwuJiogiDAi_zs1LqeN7f5ZsXbFtXGqGc0Pd
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

### Transaction Escrow Approval
```bash
 go run main.go generate transaction escrow-approval --transaction-id -2546596465476625657 --approval true --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --fee 111
```

### Block Generating Fake Blocks

```
go run main.go generate block fake-blocks --numberOfBlocks=1000 --blocksmithSecretPhrase='sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness' --out='../resource/zoobc.db'
```

### Account Generating Randomly

```
go run main.go generate account random
```

### Account Generating From Seed
```bash
Flags:
      --hd            --hd allow to generate account HD (default true)
  -h, --help          help for from-seed
      --seed string   Seed that is used to generate the account

Global Flags:
      --signature-type int32   signature-type that provide type of signature want to use to generate the account
```
Example:
```bash
go run main.go generate account from-seed --seed "concur v
ocalist rotten busload gap quote stinging undiluted surfer go
ofiness deviation starved"
### Genesis Generate From cmd/genesisblock/preRegisteredNodes.json

```

### Account Generating multisig
```bash
go run main.go generate account multisig --addresses BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN --addresses BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7 --addresses BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J —-min-sigs=2 --nonce=3
```

### Account Generate with spesific signature type
```
go run main.go generate account random  --signature-type 1
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

```
