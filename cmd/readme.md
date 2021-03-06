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

## Transaction Commands

### Transaction general flag

- `--output` to provide generated ouput type. Example: `--ouput bytes`
- `--version` to provide version of transaction. Example: `--version 1`
- `--timestamp` to provide timestamp of trasaction. Example: `--timestamp 1234567`
- `--sender-seed` to provide the seed of sender transaction. Example: `--sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"`
- `--sender-address` transaction's sender address
- `--recipient` provide recepient transaction. Example `--recipient VZvYd80p5S-rxSNQmMZwYXC7LyAzBmcfcj4MUUAdudWM`
- `--fee` to provide fee transaction, Example `--fee 1`
- `--post` to define automate post transaction or not. Example: `-post true`
- `--post-host` to provide where the transaction will post. Example: `--post-host "127.0.0.1:7000"`
- `--message` include an arbitrary message in the transaction (max 256 bytes). Example: `--message "test message"`

### Transaction Send ZBC

```
go run main.go transaction send-zbc --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --recipient ZBC_3WWDF4S2_IZVG2HHD_VOPSCNGN_COLYZ2OZ_M4QJZ4OL_44YHTKVC_2TPZBZAU --amount 5000000000
```

### Transaction send zbc escrow, set flag `--escrow true` and 3 more fields: `--approver-address`, `--timeout`, `--commission` and `--instruction`

```bash
go run main.go transaction send-zbc --escrow true --approver-address ZBC_EQAR73E3_JYON3222_KL5YYC54_Y54KMULE_KF4YSXUS_DR72QQNX_FOO2KA3L --timeout 200 --sender-seed "execute beach inflict session course dance vanish cover lawsuit earth casino fringe waste warfare also habit skull donate window cannon scene salute dawn good" --amount 1111 --commission 111 --instruction "Check amount should be 111" --recipient ZBC_3WWDF4S2_IZVG2HHD_VOPSCNGN_COLYZ2OZ_M4QJZ4OL_44YHTKVC_2TPZBZAU
```

### Transaction Register Node

```bash
go run main.go transaction register-node --node-owner-account-address="ZBC_EQAR73E3_JYON3222_KL5YYC54_Y54KMULE_KF4YSXUS_DR72QQNX_FOO2KA3L" --sender-address="ZBC_EQAR73E3_JYON3222_KL5YYC54_Y54KMULE_KF4YSXUS_DR72QQNX_FOO2KA3L" --recipient="ZBC_3WWDF4S2_IZVG2HHD_VOPSCNGN_COLYZ2OZ_M4QJZ4OL_44YHTKVC_2TPZBZAU" --sender-seed="pizza midnight enforce morning panda entire fatal plastic piece roof edge subject soap taxi car swim cash future egg burden stone step fire miracle" --node-seed="armed midnight enforce morning panda entire fatal plastic piece roof edge subject soap taxi car swim cash future egg burden stone step fire pizza" --locked-balance=100000000
```

### Transaction Update Node Registration

```bash
go run main.go transaction update-node --timestamp 1257894000 --node-owner-account-address="ZBC_EQAR73E3_JYON3222_KL5YYC54_Y54KMULE_KF4YSXUS_DR72QQNX_FOO2KA3L" --sender-address="ZBC_EQAR73E3_JYON3222_KL5YYC54_Y54KMULE_KF4YSXUS_DR72QQNX_FOO2KA3L" --recipient="ZBC_3WWDF4S2_IZVG2HHD_VOPSCNGN_COLYZ2OZ_M4QJZ4OL_44YHTKVC_2TPZBZAU" --sender-seed="pizza midnight enforce morning panda entire fatal plastic piece roof edge subject soap taxi car swim cash future egg burden stone step fire miracle" --node-seed="armed midnight enforce morning panda entire fatal plastic piece roof edge subject soap taxi car swim cash future egg burden stone step fire pizza" --locked-balance=100000000
```

### Transaction Claim Node

```bash
go run main.go transaction claim-node --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --node-owner-account-address "ZBC_EQAR73E3_JYON3222_KL5YYC54_Y54KMULE_KF4YSXUS_DR72QQNX_FOO2KA3L" --node-seed "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness" --recipient "ZBC_3WWDF4S2_IZVG2HHD_VOPSCNGN_COLYZ2OZ_M4QJZ4OL_44YHTKVC_2TPZBZAU" --poow-hex "7233537248687a792d35726c71475f644f473258626a504263574f68445552495070465267675254732d327458d880d3d1e6d68a8afeaa2c030ce50b7562fca7b7cb2ddac419c6e2ee33e0a7030000004d4e8d33954aa3deee656de56289e77d17ba29baff32da82147500e354ceaacf6cdafd6437a1037f243574dbeb2b81f52dd459ae8f0ee2ce4cbc272f832"
```

### Transaction Remove Node

```bash
go run main.go transaction remove-node --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"  --node-seed "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness"
```

### Transaction Set Account Dataset

```bash
go run main.go transaction set-account-dataset --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --recipient "ZBC_3WWDF4S2_IZVG2HHD_VOPSCNGN_COLYZ2OZ_M4QJZ4OL_44YHTKVC_2TPZBZAU" --property "Member" --value "Welcome to the jungle"
```

### Transaction Remove Account Dataset

```bash
go run main.go transaction remove-account-dataset --timestamp 1257894000 --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --recipient "ZBC_3WWDF4S2_IZVG2HHD_VOPSCNGN_COLYZ2OZ_M4QJZ4OL_44YHTKVC_2TPZBZAU" --property "Member" --value "Good Boy"
```

### Transaction Escrow Approval

```bash
 go run main.go transaction escrow-approval --transaction-id -2546596465476625657 --approval true --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --fee 111
```

### Transaction Multi Signatures

```bash
Flags:
      --address-signatures stringToString   address:signature list --address1='signature1' --address2='signature2' (default [])
      --addresses strings                   list of participants --addresses='address1,address2'
  -h, --help                                help for multi-signature
      --min-signature uint32                minimum number of signature required for the transaction to be valid
      --nonce int                           random number / access code for the multisig info
      --transaction-hash string             hash of transaction being signed by address-signature list hex
      --unsigned-transaction string         hex string of the unsigned transaction bytes
```

For the multi signature transaction let say want to send zbc with multisig account, need to do this steps:

1. Generate transaction send zbc, make sure with argument `--hash`. It will be `--unsigned-transaction` value on multi signature generator.
2. Sign the transaction to get the transaction hash, and it will be `--transcation-has` and the last the `signature-hex` will be as `--address-signatures` value on multi signature generator. <br>

So the completed comment it will be:

```bash
go run main.go transaction  multi-signature --sender-seed="execute beach inflict session course dance vanish cover lawsuit earth casino fringe waste warfare also habit skull donate window cannon scene salute dawn good" --unsigned-transaction="01000000012ba5ba5e000000002c000000486c5a4c683356636e4e6c764279576f417a584f51326a416c77464f69794f395f6e6a49336f7135596768612c000000486c38393154655446784767574f57664f4f464b59725f586468584e784f384a4b38576e4d4a56366738614c41420f0000000000080000000600000000000000000000000000000000000000000000000000000000000000" --transaction-hash="21ddbdada9903da81bf17dba6569ff7e2665fec38760c7f6636419ee30da65b0" --address-signatures="HlZLh3VcnNlvByWoAzXOQ2jAlwFOiyO9_njI3oq5Ygha=00000000b4efe21822c9d63818d8d19f6c608d917b2237426d1157b4e6689b22ce6c256ccf8ec8e2c1016ab09dd4ef2b01191fe2df70b7a123fec7115d7afd5a938f9b0a"
```

### Transaction Fee Vote Commitment Vote

```bash
 go run main.go transaction fee-vote-commit --sender-seed "execute beach inflict session course dance vanish cover lawsuit earth casino fringe waste warfare also habit skull donate window cannon scene salute dawn good" -f 10
```

### Transaction Fee Vote Reveal Vote

```bash
go run main.go transaction fee-vote-reveal -f 5 -b 4 --sender-seed "execute beach inflict session course dance vanish cover lawsuit earth casino fringe waste warfare also habit skull donate window cannon scene salute dawn good"
```

## Block Commands

### Block Generating Fake Blocks

```bash
go run main.go generate block fake-blocks --numberOfBlocks=1000 --blocksmithSecretPhrase='sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness' --out='../resource/zoobc.db'
```

## Account Commands

### Account Generate Using Ed25519 Algorithm

```bash
go run main.go account ed25519 --seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --use-slip10
```

### Account Generate Using Bitcoin Algorithm

```bash
go run main.go account bitcoin --seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --private-key-length 32 --public-key-format 1
### Genesis Generate From cmd/genesisblock/preRegisteredNodes.json

```

### Account Generating multisig

```bash
go run main.go account multisig --addresses "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN" --addresses "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7" --addresses "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J" —-min-sigs=2 --nonce=3
```

### Account Converting from encoded to hex

```bash
go run main.go account hexconv --encodedAccountAddress="ZBC_3WWDF4S2_IZVG2HHD_VOPSCNGN_COLYZ2OZ_M4QJZ4OL_44YHTKVC_2TPZBZAU" --accountType=0
```

### Account Decoding and hex-encoded full account address to its human readable encoded format

```bash
go run main.go account account hexdecode --hexAccountAddress="00000000e1e6ea65267121801089048c3a1dd863aea1fab123977677c612658a749a8a01"
```

### Account Generating account_address table by parsing account_balance table of a given zoobc.db (for debug purposes only)

```bash
go run main.go account generateaddresstable
```

## Other Commands

### Wallet certificate decryption (aid to Genesis generation)

```bash
Usage:
  zoobc decryptcert
```

or

```bash
go run main.go decryptcert
```

The command outputs resource/generated/decrypted/hosted_preRegisteredNodes.json file containing all decrypted certificates in a form that
can be
easily imported for 'genesis generate' command (see command 'Genesis' below)

note:
this command requires this specific input file:

```bash
resource/templates/certificates.json
```

with this structure (the one below is an example):

```json
[
  {
    "encryptedCert": "U2FsdGVkX18gZYg7TxccQYcSbs5Q4ToyFfD1d7ROI85lz8zka5N9FW0StDo3OXckj3Nyq9El+f9s68+F328R/fB4MxCpBJ/8uInt4sLY67dz1ps8trFmAowXYxT/gCjQqaFOttrfYXOhVteiOOV0pM+G9vZzDQ+GuwZFkMI+zqE4LlE/Do4WvWMaKofMiHlqBMzsTvLSG17o6k4VnvkSNAbpbxzaR8KE6iqzjFgB2xiZEMjeWJ9BgCODVrY+mAopVd1sL0aZ9Ya/Y0ZaVZ0Kiw==",
    "password": "123123"
  }
]
```

### Generate new Genesis

```bash
Usage:
  zoobc genesis generate [flags]

Flags:
      --applicationCodeName string      application code name (default "ZBC_main")
      --applicationVersion string       application code version (default "1.0.0")
  -f, --dbPath string                   path of blockchain's database to be used as data source in case the -w flag is used. If not set, the default resource folder is used (default "../resource/")
      --deploymentName string           nomad task name associated to this deployment (default "zoobc-alpha")
  -e, --env-target string               env mode indeed a.k.a develop,staging,alpha (default "alpha")
  -n, --extraNodes int                  number of 'extra' autogenerated nodes to be deployed using cluster_config.json
  -h, --help                            help for generate
      --kvFileCustomConfigFile string   (optional) full path (path + fileName) of a custom cluster_config.json file to use to generate consulKvInitScript.sh instead of the automatically generated in resource/generated/genesis directory
      --logLevels string                default log levels for all nodes (for kvConsulScript.sh). example: 'warn info fatal error panic' (default "fatal error panic")
  -o, --output string                   output generated files target (default "resource")
  -t, --timestamp int                   genesis timestamp, in unix epoch time, with resolution in seconds (default 1596708000)
      --wellKnownPeers string           default wellKnownPeers for all nodes (for kvConsulScript.sh). example: 'n0.alpha.proofofparticipation.network n1.alpha.proofofparticipation.network n2.alpha.proofofparticipation.network' (default "127.0.0.1:8001")
```

```bash
(from cmd directory)
go run main.go genesis generate -e {local,staging,develop,alpa} -o dist

# full example used for beta network
genesis generate --deploymentName=zoobc-beta -e=beta --applicationVersion=1.0.0 --applicationCodeName=Zoobc-beta --timestamp=1600678800 --wellKnownPeers=[139.162.126.21:8002 172.104.117.98:8002 139.162.71.117:8002 139.162.116.75:8002 172.105.211.220:8002 139.162.85.202:8002 139.162.27.172:8002 139.162.4.186:8002 172.105.166.14:8002 172.105.185.12:8002 172.105.18.138:8002 172.105.23.55:8002 172.105.104.251:8002 172.105.11.141:8002 172.105.22.227:8002 172.105.18.50:8002 139.162.154.148:8002 172.104.237.116:8002 139.162.162.254:8002 172.105.248.206:8002]
```

Starting from a 'seatSale.json' and/or a 'preRegisteredNodes.json', to be placed in resource/templates directory, it will generate files
such as genesis.go, consulKVInit.sh consul script and more, you can check these inside `${-o}/generated/genesis` directory.

structure of the two input files is:

seatSale.json

- this are all nodes in Ethereum contract used for the seat sale

```json
[
  {
    "AccountAddress": "ZBC_Y4CQBB4J_E2L6TWL5_VI2B3OB4_B3ECV5E5_6DGRQW3Q_VYHUTAAV_L2TNZYKC",
    "NodePublicKey": "ZNK_OGG7WCAZ_X5QJVGOC_JM3ND4ZG_ZLX3IQDT_WSMNXBA7_CLI5DJCB_73EDKLLD",
    "Smithing": true
  }
]
```

preRegisteredNodes.json

- this are all nodes hosted by BlockChainZoo and contains also the node seeds
- this is a subset of seatSale.json (Ethereum contract contains all notes in the registry at genesis)

```json
[
  {
    "NodePublicKey": "ZNK_W42TC2NL_XENICD3A_RAL65XGK_VZTIAN7T_IPGXAH2L_E4O2RVEN_YYUEX3XS",
    "NodeSeed": "a mnemonic passphrase",
    "AccountAddress": "ZBC_S7ORQQER_ME3GWQOS_WP5TCSQ7_M6AVTQWB_N7CU2Q3Y_L4BQI5IH_YKCICHYY",
    "Smithing": true
  }
]
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

### Scrambled Nodes

```
go run main.go scrambledNodes --db-name zoobc_2.db --height 0
```

### Priority Peers

```
go run main.go generate priorityPeers --db-name zoobc_2.db --height 11153 --sender-full-address "n56.alpha.proofofparticipation.network:8001"
```

### Transaction Liquid Payment

```
go run main.go transaction liquid-payment --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --recipient omtchrWztGDKzBftKfEarsed913s41ReV7qpMOHsFdC8 --amount 5000000000 --complete-minutes 3
```

### Transaction Liquid Payment Stop

```
go run main.go transaction liquid-payment-stop --sender-seed "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved" --transaction-id "4032174520571320308"
```

## Snapshot

Snapshot command aim to generate new snapshot files, and also import snapshot, get payload and store payload into database. This command for developer who want to test integration of snapshot is working well or not.
There are sub commands:

1.  New<br>
    Aim to generate new snapshot files, based on latest state of block chain, and store manifest into database, actually will stored to new database named `dump.db` same path with snapshot path target. if you want store to the real database just set `--dump false`.

    ````bash
    Snapshot sub command that aim to generating new snapshot file based on database target

        Usage:
          zoobc snapshot new [flags]

        Flags:
          -b, --height uint32   Block height target to snapshot
          -h, --help            help for new

        Global Flags:
          -n, --db-name string   Database name target (default "zoobc.db")
          -p, --db-path string   Database path target (default "resource")
          -d, --dump             Dump result out (default true)
          -f, --file string      Snapshot file location (default "resource/snapshot")

        ```

    ````

2.  Import
    Aim to import payload from snapshot files and will store into database, actually will store into `dump.db` as default which if `dump.db` is available, better do `snpashot new` before doing this command.

    ```bash
    Snapshot sub command simulation for import from snapshot file and storing snapshot payload into a database target

        Usage:
          zoobc snapshot import [flags]

        Flags:
          -h, --help   help for import

        Global Flags:
          -n, --db-name string   Database name target (default "zoobc.db")
          -p, --db-path string   Database path target (default "resource")
          -d, --dump             Dump result out (default true)
          -f, --file string      Snapshot file location (default "resource/snapshot")
    ```
