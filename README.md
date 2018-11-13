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
* `--database`: (default store.db) Specify where to load database of the disk-based key/value store from. The database is created if it does not exist yet.
* `--address`: (default: localhost:8000) Specify starting address and port, in format `IP:PORT`
* `--bootstrap`: (default: localhost:8000) Specify the address and port of the boostrapping node. Note that when this option is not specified, the miner connects to itself.
* `--wallet`: (default: wallet.txt) Load the public key from this file. A new private key is generated if it does not exist yet. Note that only the public key is required.
* `--multisig`: (optional) The file to load the multisig's private key from.
* `--commitment`: The file to load the validator's commitment key from (will be created if it does not exist)
* `--rootkey`: (default: key.txt) The file to load root's public key from this file. A new public private key is generated if it does not exist yet. Note that only the public key is required.
* `--rootcommitment`: The file to load root's commitment key from. A new commitment key is generated if it does not exist yet.
* `--confirm`: In order to review the miner startup options, the user must press Enter before the miner starts.

Example

Using a sample scenario, the use of the command line should become clear.

Let's assume we want to start two miners, miner `A` and miner `B`, whereas miner `A` acts as the bootstrap node.
Further assume that we start from scratch and no key files have been created yet.

Miner A (Root)
* Database: `StoreA.db`
* Address: `localhost:8000`
* Bootstrap Address: `localhost:8000`
* Wallet: `WalletA.txt`
* Commitment: `CommitmentA.txt`
* Root Wallet: `WalletA.txt`
* Root Commitment: `CommitmentA.txt`


Miner B
* Database: `StoreB.db`
* Address: `localhost:8001`
* Bootstrap Address: `localhost:8000`
* Wallet: `WalletB.txt`
* Commitment: `CommitmentB.txt`

Commands

```bash
bazo-miner start --database StoreA.db --address localhost:8000 --bootstrap localhost:8000 --wallet WalletA.txt --commitment CommitmentA.txt --multisig WalletA.txt --rootwallet WalletA.txt --rootcommitment CommitmentA.txt
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

Start miner B, using the generated `WalletB.txt` and `CommitmentB.txt` (e.g. copy the files to the Bazo miner directory):

```bash
bazo-miner start --database StoreB.db --address localhost:8001 --bootstrap localhost:8000 --wallet WalletB.txt --commitment CommitmentB.txt --rootwallet WalletA.txt --rootcommitment CommitmentA.txt
```

Note that both files specified for `--rootwallet` and `--rootcommitment` only require to contain the wallet and commitemt public key respectively.

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

