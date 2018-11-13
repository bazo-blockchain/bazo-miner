package main

import (
	"github.com/bazo-blockchain/bazo-miner/cli"
	"github.com/bazo-blockchain/bazo-miner/storage"
	cli2 "github.com/urfave/cli"
	"os"
)

func main() {
	logger := storage.InitLogger()

	app := cli2.NewApp()

	app.Name = "bazo-miner"
	app.Usage = "the command line interface for running a full Bazo blockchain node implemented in Go."
	app.Version = "1.0.0"
	app.EnableBashCompletion = true
	app.Commands = []cli2.Command {
		cli.GetStartCommand(logger),
		cli.GetGenerateWalletCommand(),
		cli.GetGenerateCommitmentCommand(),
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}
