# bazo-miner
`bazo-miner` is the the command line interface for running a full Bazo blockchain node implemented in Go.

[![Build Status](https://travis-ci.org/bazo-blockchain/bazo-miner.svg?branch=master)](https://travis-ci.org/bazo-blockchain/bazo-miner)

## Setup Instructions

The programming language Go (developed and tested with version >= 1.9) must be installed, the properties $GOROOT and $GOPATH must be set. For more information, please check out the [official documentation](https://github.com/golang/go/wiki/SettingGOPATH).

### Prerequisites

Before the bazo-miner can be started, two public-private key-pairs are required. The key-pairs can be generated with the bazo-keypairgen application. Run the following instructions in your terminal.

1. Download the bazo-keypairgen application.
```
go get github.com/bazo-blockchain/bazo-keypairgen
```

2. Navigate to the bazo-keypairgen directory and build the application.
```
cd ~/go/src/github.com/bazo-blockchain/bazo-keypairgen
go build
```

Note: Replace `~/go` with your `$GOPATH`.

3. Run the application to generate the _validator_ public-private keypair. The validator is the keyﬁle's name containing the validator's public key.
```
./bazo-keypairgen validator.txt
```

4. Run the application to generate the _multisig_ public-private keypair. The multisig is the keyﬁle's name containing the multi-signature server's public key.
```
./bazo-keypairgen multisig.txt
```

### Getting Started

1. Download the bazo-miner application.
```
go get github.com/bazo-blockchain/bazo-miner
```

2. Copy both previously generated files `validator.txt` and `multisig.txt` into the root folder of the bazo-miner folder.

3. Open the storage configuration file `storage.configs.go` in an editor of your choice.
```
$GOPATH/src/github.com/bazo-blockchain/bazo-miner/storage/configs.go
```

Replace the value of `INITROOTPUBKEY1` with the first line of `validator.txt`. Replace the value of `INITROOTPUBKEY2` with the second line of `validator.txt`.

4. Build the application.
```
cd ~/go/src/github.com/bazo-blockchain/bazo-miner
go build
```

5. Run the application (bootstrap node).
```
./bazo-miner "database_file.db" "localhost:8000" "localhost:8000" "validator.txt" "multisig.txt" "commitment.txt"
```

The ipport number must be preﬁxed with ":". If the miner is intended to run locally, the localhost ip address has to be passed with the ipport. Otherwise the miner tries to connect to the network. Note that "database_file.db" and "commitment.txt" are created if they do not exist.
