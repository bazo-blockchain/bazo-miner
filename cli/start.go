package cli

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/urfave/cli"
	"log"
)

type startArgs struct {
	dbname 					string

	myNodeAddress			string
	bootstrapNodeAddress	string

	keyFileName				string
	multisigFileName		string
	commitmentFileName		string

	rootKeyFileName			string
	rootCommitmentFileName	string
}

func AddStartCommand(app *cli.App, logger *log.Logger) {
	command := cli.Command		{
		Name:	"start",
		Usage:	"start the miner",
		Action:	func(c *cli.Context) error {
			args := &startArgs {
				dbname: 				c.String("database"),
				myNodeAddress: 			c.String("address"),
				bootstrapNodeAddress: 	c.String("bootstrap"),
				keyFileName: 			c.String("key"),
				multisigFileName: 		c.String("multisig"),
				commitmentFileName:		c.String("commitment"),

				rootKeyFileName:		c.String("rkey"),
				rootCommitmentFileName: c.String("rcommitment"),
			err := args.ValidateInput()
			if err != nil {
				return err
			}
			fmt.Println(args.String())

			if c.Bool("confirm") {
				fmt.Scanf("\n")
			}

			return Start(args, logger)
		},
		Flags:	[]cli.Flag {
			cli.StringFlag {
				Name: 	"database, d",
				Usage: 	"Load database of the disk-based key/value store from `FILE`",
				Value:	"store.db",
			},
			cli.StringFlag {
				Name: 	"address, a",
				Usage: 	"Start node at `IP:PORT`",
				Value: 	"localhost:8000",
			},
			cli.StringFlag {
				Name: 	"bootstrap, b",
				Usage: 	"Connect to bootstrap node at `IP:PORT`",
				Value: 	"localhost:8000",
			},
			cli.StringFlag {
				Name: 	"key, k",
				Usage: 	"Load validator's public key from `FILE`",
				Value: 	"key.txt",
			},
			cli.StringFlag {
				Name: 	"multisig, m",
				Usage: 	"Load multi-signature server’s public key from `FILE`",
				Value: 	"multisig.txt",
			},
			cli.StringFlag {
				Name: 	"commitment, c",
				Usage: 	"Load validator's RSA public-private key from `FILE`",
				Value: 	"commitment.txt",
			},
			cli.StringFlag {
				Name: 	"rkey",
				Usage: 	"Load root's public key from `FILE`",
				Value: 	"key.txt",
			},
			cli.StringFlag {
				Name: 	"rcommitment",
				Usage: 	"Load root's RSA public-private key from `FILE`",
				Value: 	"commitment.txt",
			},
			cli.BoolFlag {
				Name: 	"confirm",
				Usage: 	"User must confirm start parameters before running",
			},
		},
	}

	app.Commands = append(app.Commands, command)
}

func Start(args *startArgs, logger *log.Logger) error {
	storage.Init(args.dbname, args.bootstrapNodeAddress)
	p2p.Init(args.myNodeAddress)

	validatorPubKey, err := crypto.ExtractECDSAPublicKeyFromFile(args.keyFileName)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	multisigPubKey, err := crypto.ExtractECDSAPublicKeyFromFile(args.multisigFileName)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	commPrivKey, err := crypto.ExtractRSAKeyFromFile(args.commitmentFileName)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	rootPubKey, err := crypto.ExtractECDSAPublicKeyFromFile(args.rootKeyFileName)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	rootCommPrivKey, err := crypto.ExtractRSAKeyFromFile(args.rootCommitmentFileName)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	miner.Init(validatorPubKey, multisigPubKey, rootPubKey, commPrivKey, rootCommPrivKey)
	return nil
}

func (args startArgs) ValidateInput() error {
	if len(args.dbname) == 0 {
		return errors.New("argument missing: dbname")
	}

	if len(args.myNodeAddress) == 0 {
		return errors.New("argument missing: myNodeAddress")
	}

	if len(args.bootstrapNodeAddress) == 0 {
		return errors.New("argument missing: bootstrapNodeAddress")
	}

	if len(args.keyFileName) == 0 {
		return errors.New("argument missing: keyFileName")
	}

	if len(args.multisigFileName) == 0 {
		return errors.New("argument missing: multisigFileName")
	}

	if len(args.commitmentFileName) == 0 {
		return errors.New("argument missing: commitmentFileName")
	}

	if len(args.rootKeyFileName) == 0 {
		return errors.New("argument missing: rootKeyFileName")
	}

	if len(args.rootCommitmentFileName) == 0 {
		return errors.New("argument missing: rootCommitmentFileName")
	}

	return nil
}

func (args startArgs) String() string {
	return fmt.Sprintf("Starting bazo miner with arguments \n" +
			"- Database Name:\t\t %v\n" +
			"- My Address:\t\t\t %v\n" +
			"- Bootstrap Address:\t\t %v\n" +
			"- Key Filename:\t\t\t %v\n" +
			"- Multisig Filename:\t\t %v\n" +
			"- Commitment Filename:\t\t %v\n" +
			"- Root Key Filename:\t\t %v\n" +
			"- Root Commitment Filename:\t %v\n",
		args.dbname,
		args.myNodeAddress,
		args.bootstrapNodeAddress,
		args.keyFileName,
		args.multisigFileName,
		args.commitmentFileName,
		args.rootKeyFileName,
		args.rootCommitmentFileName)
}