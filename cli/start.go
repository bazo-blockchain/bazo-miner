package cli

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"log"
)

type startArgs struct {
	dbname 					string
	myNodeAddress			string
	bootstrapNodeAddress	string
	keyFileName				string
	multisigFile			string
	commitmentFile			string
	rootKeyFile				string
	rootCommitmentFile		string
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
				multisigFile: 			c.String("multisig"),
				commitmentFile:			c.String("commitment"),
				rootKeyFile:			c.String("rootkey"),
				rootCommitmentFile: 	c.String("rootcommitment"),
			}

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
				Usage: 	"start node at `IP:PORT`",
				Value: 	"localhost:8000",
			},
			cli.StringFlag {
				Name: 	"bootstrap, b",
				Usage: 	"connect to bootstrap node at `IP:PORT`",
				Value: 	"localhost:8000",
			},
			cli.StringFlag {
				Name: 	"key, k",
				Usage: 	"load validator's public key from `FILE`",
				Value: 	"key.txt",
			},
			cli.StringFlag {
				Name: 	"multisig, m",
				Usage: 	"load multi-signature server’s public key from `FILE`",
				Value: 	"multisig.txt",
			},
			cli.StringFlag {
				Name: 	"commitment, c",
				Usage: 	"load validator's RSA public-private key from `FILE`",
				Value: 	"commitment.txt",
			},
			cli.StringFlag {
				Name: 	"rootkey",
				Usage: 	"load root's public key from `FILE`",
				Value: 	"key.txt",
			},
			cli.StringFlag {
				Name: 	"rootcommitment",
				Usage: 	"load root's RSA public-private key from `FILE`",
				Value: 	"commitment.txt",
			},
			cli.BoolFlag {
				Name: 	"confirm",
				Usage: 	"user must press enter before starting the miner",
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

	multisigPubKey, err := crypto.ExtractECDSAPublicKeyFromFile(args.multisigFile)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	commPrivKey, err := crypto.ExtractRSAKeyFromFile(args.commitmentFile)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	rootPubKey, err := crypto.ExtractECDSAPublicKeyFromFile(args.rootKeyFile)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	rootCommPrivKey, err := crypto.ExtractRSAKeyFromFile(args.rootCommitmentFile)
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

	if len(args.keyFile) == 0 {
		return errors.New("argument missing: keyFile")
	}

	if len(args.multisigFile) == 0 {
		return errors.New("argument missing: multisigFile")
	}

	if len(args.commitmentFile) == 0 {
		return errors.New("argument missing: commitmentFile")
	}

	if len(args.rootKeyFile) == 0 {
		return errors.New("argument missing: rootKeyFile")
	}

	if len(args.rootCommitmentFile) == 0 {
		return errors.New("argument missing: rootCommitmentFile")
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
		args.multisigFile,
		args.commitmentFile,
		args.rootKeyFile,
		args.rootCommitmentFile)
}