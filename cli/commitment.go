package cli

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/urfave/cli"
)

func GetGenerateCommitmentCommand() cli.Command {
	return cli.Command {
		Name:	"generate-commitment",
		Usage:	"generate a new pair of commitment keys",
		Action:	func(c *cli.Context) error {
			filename := c.String("file")
			privKey, err := crypto.ExtractRSAKeyFromFile(filename)

			fmt.Printf("Commitment generated successfully.\n")
			fmt.Printf("PubKeyE: %x\n", privKey.PublicKey.E)
			fmt.Printf("PubKeyN: %x\n", privKey.PublicKey.N)
			fmt.Printf("PrivKey: %x\n", privKey.D)

			return err
		},
		Flags:	[]cli.Flag {
			cli.StringFlag {
				Name: 	"file",
				Usage: 	"the new commitment key's `FILE` name",
			},
		},
	}
}
