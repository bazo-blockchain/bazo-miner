package main

import (
	"github.com/bazo-blockchain/bazo-miner/cli"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	cli2 "github.com/urfave/cli"
	"os"
)

func main() {
	Trie, _ := trie.New(common.Hash{}, trie.NewDatabase(ethdb.NewMemDatabase()))
	println(Trie)

	// test commit for new sharding branch
	app := cli2.NewApp()

	app.Name = "bazo-miner"
	app.Usage = "the command line interface for running a full Bazo blockchain node implemented in Go."
	app.Version = "1.0.0"
	app.EnableBashCompletion = true
	app.Commands = []cli2.Command {
		cli.GetStartCommand(),
		cli.GetGenerateWalletCommand(),
		cli.GetGenerateCommitmentCommand(),
	}

	err := app.Run(os.Args)
	if err != nil {
		logger := storage.InitLogger()
		logger.Fatal(err)
	}
}
