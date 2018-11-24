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
* Data Directory: `NodeA`, contains `wallet.key`, `commitment.key`, `multisig.key` and `store.db`
* Address: `localhost:8000`
* Bootstrap Address: `localhost:8000`


Miner B
* Data Directory: `NodeB`, contains `wallet.key`, `commitment.key`, `multisig.key` and `store.db`
* Address: `localhost:8001`
* Bootstrap Address: `localhost:8000`

Commands

```bash
./bazo-miner start --dataDir NodeA --address localhost:8000 --bootstrap localhost:8000
```

We start miner A at address and port `localhost:8000` and connect to itself by setting the bootstrap address to the same address.
Note that we could have omitted these two options since they are passed by default with these values.
Wallet and commitment keys are automatically created. Using this command, we define miner A as the root.

In a second terminal, run

```bash
./bazo-miner start --dataDir NodeB --address localhost:8001 --bootstrap localhost:8000
```

Notice how miner B ist started at address and port `localhost:8001` but bootstraps to miner A.
Running this command will give you an error message, i.e.,

```bash
Acc (...) not in the state.
```

Starting miner B requires more work since every miner must have sufficient funds and be part of the set of validators.
The minimum amount of coins required for staking is defined in the configuration of Bazo.

Our current Bazo miner directory should look like this:

```
bazo-miner (root folder)
-- NodeA
---- wallet.key
---- commitment.key
---- multisig.key
---- store.db
-- NodeB
---- wallet.key
---- commitment.key
---- store.db
-- bazo-miner (executable)
``` 

In our case, we can use the wallet of miner A wallet to move funds to miner B. 
* Copy `wallet.key` and `multisig.key` from the directory `NodeA` to the Bazo client directory and rename the file to `WalletA.key`.
* Copy `wallet.key` and `commitment.key` from the directory `NodeB` to the Bazo client directory and rename the file to `WalletB.key` and `CommitmentB.key` respectively.

Using the [Bazo client](https://github.com/bazo-blockchain/bazo-client), we transfer 2000 coins from A to B:

```bash
./bazo-client funds --from WalletA.key --to WalletB.key --txcount 0 --amount 2000 --multisig Multisig.key 
```

Check the terminal of miner B. The error message should change to (may need some time until the FundsTx is validated)

```bash
Validator (...) is not part of the validator set.
```

Now, miner B has to join the pool of validators (enable staking):

```bash
./bazo-client staking enable --wallet WalletB.key --commitment CommitmentB.key
```

Again, check miner B's terminal. After some time, miner B should validate and create blocks automatically.

### Generate a wallet

Generate a new public and private wallet keypair.

```bash
bazo-miner generate-wallet [command options] [arguments...]
```

Options
* `--file`: Save the public private wallet keypair to this file.

Example

```bash
./bazo-miner generate-wallet --file wallet.txt
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
./bazo-miner generate-commitment --file commitment.txt
```

