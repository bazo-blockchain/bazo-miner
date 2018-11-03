package main

import (
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/urfave/cli"
	"os"
)

func main() {

	logger := storage.InitLogger()

	app := cli.NewApp()

	// Global app config
	app.Name = "bazo-miner"
	app.Usage = "the command line interface for running a full Bazo blockchain node implemented in Go."
	app.Version = "1.0.0"
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag {
		cli.StringFlag {
			Name: 	"database",
			Usage: 	"Load database of the disk-based key/value store from `FILE`",
			Value:	"keystore/store.db",
		},
		cli.StringFlag {
			Name: 	"address",
			Usage: 	"Start node at `IP:PORT`",
			Value: 	"localhost:8000",
		},
		cli.StringFlag {
			Name: 	"bootstrap",
			Usage: 	"Connect to bootstrap node at `IP:PORT`",
			Value: 	"localhost:8000",
		},
		cli.StringFlag {
			Name: 	"key",
			Usage: 	"Load validator's public key from `FILE`",
			Value: 	"keystore/validator.txt",
		},
		cli.StringFlag {
			Name: 	"multisig",
			Usage: 	"Load multi-signature server’s public key from `FILE`",
			Value: 	"keystore/multisig.txt",
		},
		cli.StringFlag {
			Name: 	"commitment",
			Usage: 	"Load RSA public-private key from `FILE`",
			Value: 	"keystore/commitment.txt",
		},
	}

	app.Action = func(c *cli.Context) error {
		dbname := c.String("database")
		thisIpport := c.String("address")
		bootstrapIpport := c.String("bootstrap")
		validatorFileName := c.String("key")
		multisigFileName := c.String("multisig")
		commFileName := c.String("commitment")

		storage.Init(dbname, bootstrapIpport)
		p2p.Init(thisIpport)

		validatorPrivKey, err := crypto.ExtractECDSAKeyFromFile(validatorFileName)
		if err != nil {
			logger.Printf("%v\n", err)
			return err
		}

		multisigPrivKey, err := crypto.ExtractECDSAKeyFromFile(multisigFileName)
		if err != nil {
			logger.Printf("%v\n", err)
			return err
		}

		commPrivKey, err := crypto.ExtractRSAKeyFromFile(commFileName)
		if err != nil {
			logger.Printf("%v\n", err)
			return err
		}

		miner.Init(&validatorPrivKey.PublicKey, &multisigPrivKey.PublicKey, &commPrivKey)
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}
