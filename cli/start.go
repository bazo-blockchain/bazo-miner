package cli

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"log"
	"os"
)

type startArgs struct {
	dataDirectory			string
	myNodeAddress			string
	bootstrapNodeAddress	string
	multisigFile			string
	rootKeyFile				string
	rootCommitmentFile		string
}

func GetStartCommand(logger *log.Logger) cli.Command {
	return cli.Command {
		Name:	"start",
		Usage:	"start the miner",
		Action:	func(c *cli.Context) error {
			args := &startArgs {
				dataDirectory: 			c.String("dataDir"),
				myNodeAddress: 			c.String("address"),
				bootstrapNodeAddress: 	c.String("bootstrap"),
				multisigFile: 			c.String("multisig"),
				rootKeyFile:			c.String("rootwallet"),
				rootCommitmentFile: 	c.String("rootcommitment"),
			}

			if !c.IsSet("bootstrap") {
				args.bootstrapNodeAddress = args.myNodeAddress
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
				Name: 	"dataDir, d",
				Usage: 	"Data directory for the database and keystore",
				Value:	"bazodata",
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
				Name: 	"multisig, m",
				Usage: 	"Load multi-signature server’s public key from `FILE`",
			},
			cli.StringFlag {
				Name: 	"rootwallet",
				Usage: 	"Load root's public key from `FILE`",
				Value: 	"rootwallet.txt",
			},
			cli.StringFlag {
				Name: 	"rootcommitment",
				Usage: 	"Load root's RSA public-private key from `FILE`",
				Value: 	"rootcommitment.txt",
			},
			cli.BoolFlag {
				Name: 	"confirm",
				Usage: 	"User must press enter before starting the miner",
			},
		},
	}
}

func Start(args *startArgs, logger *log.Logger) error {
	if _, err := os.Stat(args.dataDirectory); os.IsNotExist(err) {
		err = os.MkdirAll(args.dataDirectory, 0755)
		if err != nil {
			return err
		}
	}

	const (
		database	= "store.db"
		wallet 		= "wallet.key"
		commitment 	= "commitment.key"
	)

	storage.Init(args.dataDirectory + "/" + database, args.bootstrapNodeAddress)
	p2p.Init(args.myNodeAddress)

	validatorPubKey, err := crypto.ExtractECDSAPublicKeyFromFile(args.dataDirectory + "/" + wallet)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	rootPrivKey, err := crypto.ExtractECDSAKeyFromFile(args.rootKeyFile)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	var multisigPubKey *ecdsa.PublicKey
	if len(args.multisigFile) > 0 {
		multisigPubKey, err = crypto.ExtractECDSAPublicKeyFromFile(args.multisigFile)
		if err != nil {
			logger.Printf("%v\n", err)
			return err
		}
	} else {
		multisigPubKey = &rootPrivKey.PublicKey
	}

	commPrivKey, err := crypto.ExtractRSAKeyFromFile(args.dataDirectory + "/" + commitment)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	rootCommPrivKey, err := crypto.ExtractRSAKeyFromFile(args.rootCommitmentFile)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	miner.Init(validatorPubKey, multisigPubKey, &rootPrivKey.PublicKey, commPrivKey, rootCommPrivKey)
	return nil
}

func (args startArgs) ValidateInput() error {
	if len(args.dataDirectory) == 0 {
		return errors.New("argument missing: dataDir")
	}

	if len(args.myNodeAddress) == 0 {
		return errors.New("argument missing: myNodeAddress")
	}

	if len(args.bootstrapNodeAddress) == 0 {
		return errors.New("argument missing: bootstrapNodeAddress")
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
			"- My Address:\t\t\t %v\n" +
			"- Bootstrap Address:\t\t %v\n" +
			"- Store Directory:\t\t\t %v\n" +
			"- Multisig File:\t\t %v\n" +
			"- Root Wallet File:\t\t %v\n" +
			"- Root Commitment File:\t %v\n",
		args.myNodeAddress,
		args.bootstrapNodeAddress,
		args.dataDirectory,
		args.multisigFile,
		args.rootKeyFile,
		args.rootCommitmentFile)
}