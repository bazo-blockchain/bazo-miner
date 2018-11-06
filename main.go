package main

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
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

	app.Commands = []cli.Command {
		{
			Name:	"start",
			Usage:	"start the miner",
			Action:	func(c *cli.Context) error {
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

				miner.Init(&validatorPrivKey.PublicKey, &multisigPrivKey.PublicKey, commPrivKey)
				return nil
			},
			Flags:	[]cli.Flag {
				cli.StringFlag {
					Name: 	"database",
					Usage: 	"Load database of the disk-based key/value store from `FILE`",
					Value:	"store.db",
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
					Value: 	"validator.txt",
				},
					cli.StringFlag {
					Name: 	"multisig",
					Usage: 	"Load multi-signature server’s public key from `FILE`",
					Value: 	"multisig.txt",
				},
					cli.StringFlag {
					Name: 	"commitment",
					Usage: 	"Load RSA public-private key from `FILE`",
					Value: 	"commitment.txt",
				},
			},
		},
		{
			Name:	"generate",
			Usage:	"generate a new pair of keys",
			Action:	func(c *cli.Context) error {
				filename := c.String("filename")
				privKey, err := crypto.ExtractECDSAKeyFromFile(filename)

				fmt.Printf("Keyfile generated successfully.\n")
				fmt.Printf("PubKeyX: %x\n", privKey.PublicKey.X)
				fmt.Printf("PubKeyY: %x\n", privKey.PublicKey.Y)
				fmt.Printf("PrivKey: %x\n", privKey.D)

				return err
			},
			Flags:	[]cli.Flag {
				cli.StringFlag {
					Name: 	"filename",
					Usage: 	"The key's `FILE` name",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}
