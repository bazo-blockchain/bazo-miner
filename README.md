# bazo-miner
`bazo-miner` is the the command line interface for running a full Bazo blockchain node implemented in Go.

[![Build Status](https://travis-ci.org/bazo-blockchain/bazo-miner.svg?branch=master)](https://travis-ci.org/bazo-blockchain/bazo-miner)

## Setup Instructions

The programming language Go (developed and tested with version >= 1.9) must be installed, the properties $GOROOT and $GOPATH must be set. For more information, please check out the [official documentation](https://github.com/golang/go/wiki/SettingGOPATH).

## Getting Started

The Bazo miner provides an intuitive and beginner-friendly command line interface.

```bash
bazo-miner [global options] command [command options] [arguments...]
```

Options
* `--help, -h`: Show help 
* `--version, -v`: Print the version

### Start the miner

Start the miner with a breeze. 

```bash
bazo-miner start [command options] [arguments...]
```

Options
* `--address`: (default: localhost:8000) Specify starting address and port, in format `IP:PORT`
* `--bootstrap`: (default: localhost:8000) Specify the address and port of the boostrapping node. Note that when this option is not specified, the miner connects to itself.
* `--dataDir`: (default: bazodata) Data directory for the database (store.db) and keystore (wallet.key, commitment.key). Database and keys are generated if they do not exist yet.
* `--confirm`: In order to review the miner startup options, the user must press Enter before the miner starts.

Example

Using a sample scenario, the use of the command line should become clear.

Let's assume we want to start two miners, miner `A` and miner `B`, whereas miner `A` acts as the bootstrap node.
Further assume that we start from scratch and no key files have been created yet.

Miner A (Root)
* Data Directory: `NodeA`
* Address: `localhost:8000`
* Bootstrap Address: `localhost:8000`
* Database: `StoreA.db`
* Wallet: `WalletA.key`
* Commitment: `CommitmentA.key`
* Root Wallet: `WalletA.key`
* Root Commitment: `CommitmentA.key`


Miner B
* Data Directory: `NodeB`
* Address: `localhost:8001`
* Bootstrap Address: `localhost:8000`
* Database: `StoreB.db`
* Wallet: `WalletB.key`
* Commitment: `CommitmentB.key`
* Root Wallet: `WalletA.key` (can contain only the public key)
* Root Commitment: `CommitmentA.key` (can contain only the public key)

Commands

```bash
bazo-miner start --dataDir NodeA --address localhost:8000 --bootstrap localhost:8000
```

We start miner A at address and port `localhost:8000` and connect to itself by setting the bootstrap address to the same address.
Note that we could have omitted these two options since they are passed by default with these values.
Wallet and commitment keys are automatically created. Using this command, we define miner A as the root.

Starting miner B requires more work since new accounts have to be registered by a root account.
In our case, we can use miner's A `WalletA.txt` (e.g. copy the file to the Bazo client directory) to create and add a new account to the network.
Using the [Bazo client](https://github.com/bazo-blockchain/bazo-client), we create a new account:

```bash
bazo-client account create --rootwallet WalletA.txt --wallet WalletB.txt 
```

The minimum amount of coins required for staking is defined in the configuration of Bazo.
Thus, miner B first needs Bazo coins to start mining and we must first send coins to miner B's account.

```bash
bazo-client funds --from WalletA.txt --to WalletB.txt --txcount 0 --amount 1000 --multisig WalletA.txt
```

Then, miner B has to join the pool of validators (enable staking):
```bash
bazo-client staking enable --wallet WalletB.txt --commitment CommitmentB.txt
```

In addition to the created `NodeA` directory before (located in the Bazo miner directory), create a new directory `NodeB`, 
copy the generated files `WalletB.txt` and `CommitmentB.txt`, as well as the root wallet (in our case `WalletA.key`) 
and the root commitment (in our case `CommitmentA.key`) to the `NodeB` directory, resulting in the following directory structure:

```
:open_file_folder: bazo-miner
-- :open_file_folder: NodeA
---- :key: WalletA.key
---- :key: CommitmentA.key
---- :floppy_disk: StoreA.db
---- :key: WalletA.key
---- :key: CommitmentA.key
-- :open_file_folder: NodeB
---- :key: WalletB.key
---- :key: CommitmentB.key
---- :floppy_disk: StoreB.db
---- :key: WalletA.key
---- :key: CommitmentA.key
-- bazo-miner (executable)
``` 

Then start the miner:

```bash
bazo-miner start --dataDir NodeB --address localhost:8001 --bootstrap localhost:8000
```

We start miner B at address and port `localhost:8001` and connect to miner A (which is the boostrap node).
Wallet and commitment keys are automatically created.

### Generate a wallet

Generate a new public and private wallet keypair.

```bash
bazo-miner generate-wallet [command options] [arguments...]
```

Options
* `--file`: Save the public private wallet keypair to this file.

Example

```bash
bazo-miner generate-wallet --file wallet.txt
```


### Generate a commitment

Generate a new public and private commitment keypair.

```bash
bazo-miner generate-commitment [command options] [arguments...]
```

Options
* `--file`: Save the public private commitment keypair to this file.

Example

```bash
bazo-miner generate-commitment --file commitment.txt
```

