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

### See more help about commands
- `go run main --help` to see available commands and flags
- `go run main {command} --help` to see to see available subcommands and flags
- `go run main {command} {subcommand} --help` to see to see available subcommands and flags of subcommand

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
```bash
go run main.go generate transaction register-node --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --node-owner-account-address "VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM" --node-seed "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness" --node-address "127.0.0.1:8001" --locked-balance 1000000000 --poow-hex "7233537248687a792d35726c71475f644f473258626a504263574f68445552495070465267675254732d327458d880d3d1e6d68a8afeaa2c030ce50b7562fca7b7cb2ddac419c6e2ee33e0a7030000004d4e8d33954aa3deee656de56289e77d17ba29baff32da82147500e354ceaacf6cdafd6437a1037f243574dbeb2b81f52dd459ae8f0ee2ce4cbc272f832"
```

### Transaction Update Node Registration

```bash
go run main.go generate transaction update-node --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --node-owner-account-address VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM --node-seed "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness" --node-address "127.0.0.1:8001" --locked-balance 10050000000000 --poow-hex "7233537248687a792d35726c71475f644f473258626a504263574f68445552495070465267675254732d327458d880d3d1e6d68a8afeaa2c030ce50b7562fca7b7cb2ddac419c6e2ee33e0a7030000004d4e8d33954aa3deee656de56289e77d17ba29baff32da82147500e354ceaacf6cdafd6437a1037f243574dbeb2b81f52dd459ae8f0ee2ce4cbc272f832"
```

### Transaction Claim Node

```bash
go run main.go generate transaction claim-node --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --node-owner-account-address "VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM" --node-seed "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness" --recipient "VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM" --poow-hex "7233537248687a792d35726c71475f644f473258626a504263574f68445552495070465267675254732d327458d880d3d1e6d68a8afeaa2c030ce50b7562fca7b7cb2ddac419c6e2ee33e0a7030000004d4e8d33954aa3deee656de56289e77d17ba29baff32da82147500e354ceaacf6cdafd6437a1037f243574dbeb2b81f52dd459ae8f0ee2ce4cbc272f832"
```

### Transaction Remove Node

```bash
go run main.go generate transaction remove-node --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"  --node-seed "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness"
```

### Transaction Set Account Dataset

```bash
go run main.go generate transaction set-account-dataset --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --recipient "VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM" --property "Member" --value "Welcome to the jungle"
```

### Transaction Remove Account Dataset

```bash
go run main.go generate transaction remove-account-dataset --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --recipient "VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM" --property "Member" --value "Good Boy"
```

### Transaction Escrow Approval
```bash
 go run main.go generate transaction escrow-approval --transaction-id -2546596465476625657 --approval true --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --fee 111
```

### Block Generating Fake Blocks

```bash
go run main.go generate block fake-blocks --numberOfBlocks=1000 --blocksmithSecretPhrase='sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness' --out='../resource/zoobc.db'
```

### Account Generate Using Ed25519 Algorithm

```bash
go run main.go generate account ed25519 --seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --use-slip10
```

### Account Generate Using Bitcoin Algorithm

```bash
go run main.go generate account bitcoin --seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --private-key-length 32 --public-key-format 1
### Genesis Generate From cmd/genesisblock/preRegisteredNodes.json

```

### Account Generating multisig
```bash
go run main.go generate account multisig --addresses "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN" --addresses "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7" --addresses "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J" —-min-sigs=2 --nonce=3
```

go run main.go genesis generate
outputs cmd/genesis.go.new and cmd/cluster_config.json

```bash

### Genesis Generate From cmd/genesisblock/preRegisteredNodes.json and resource/zoobc.db

```

go run main.go genesis generate -w
outputs cmd/genesis.go.new and cmd/cluster_config.json

```bash

### Genesis Generate From cmd/genesisblock/preRegisteredNodes.json and resource/zoobc.db, plus n random nodes registrations

```

go run main.go genesis generate -w -n 10
outputs cmd/genesis.go.new and cmd/cluster_config.json

```

```

### Generate Proof of Ownership Node Registry
```bash
go run main.go generate poow --node-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"   --node-owner-account-address "VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM"  --db-node-path "../resource" --db-node-name "zoobc.db"
--output-type "hex" 
```

### Rollback Blockchain State
```bash
go run main.go rollback blockchain --to-height 10 --db-path "../resource" --db-name "zoobc.db"
```

### Signature Signing data using Ed25519
```bash
go run main.go signature sign ed25519 --data-bytes='1, 222, 54, 12, 32' --use-slip10=true
```

### Signature Verifying data
```bash
go run main.go signature verify --data-bytes='1, 222, 54, 12, 32' --signature-hex=0000000063851d61318eaf923ff72457fd9b5716db9904aacbe53eb1bc25cd8a7bf2816c61402b0c52d4324e1336bae4ea28194d6f5c531292fd266e63a293519f20c20b --account-address=WI-u0jyKMGsPHk6K7eT1Utnxc6WiKehkIEs87Zf3fIsH
```