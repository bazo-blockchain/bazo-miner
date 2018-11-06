package cli

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/urfave/cli"
)

func AddGenerateCommand(app *cli.App) {
	command := cli.Command {
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
	}

	app.Commands = append(app.Commands, command)
}
