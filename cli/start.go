package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type startArgs struct {
	dataDirectory        string
	myNodeAddress        string
	bootstrapNodeAddress string
	isRoot				 bool
}

func GetStartCommand(logger *log.Logger) cli.Command {
	return cli.Command{
		Name:  "start",
		Usage: "start the miner",
		Action: func(c *cli.Context) error {
			args := &startArgs{
				dataDirectory:        	c.String("dataDir"),
				myNodeAddress:        	c.String("address"),
				bootstrapNodeAddress: 	c.String("bootstrap"),
				isRoot: 			  	c.Bool("isRoot"),
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
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "dataDir, d",
				Usage: "Data directory for the database and keystore",
				Value: "NodeA",
			},
			cli.StringFlag{
				Name:  "address, a",
				Usage: "Start node at `IP:PORT`",
				Value: "localhost:8000",
			},
			cli.StringFlag{
				Name:  "bootstrap, b",
				Usage: "Connect to bootstrap node at `IP:PORT`",
				Value: "localhost:8000",
			},
			cli.BoolFlag{
				Name:  "isRoot",
				Usage: "Starts the node as root",
			},
			cli.BoolFlag{
				Name:  "confirm",
				Usage: "User must press enter before starting the miner",
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
		database       = "Store.db"
		wallet         = "ValidatorWallet.key"
		commitment     = "ValidatorCommitment.key"
		rootWallet     = "RootWallet.key"
		rootCommitment = "RootCommitment.key"
		multisig       = "Multisig.key"
	)

	storage.Init(args.dataDirectory+"/"+database, args.bootstrapNodeAddress)
	p2p.Init(args.myNodeAddress)

	validatorPubKey, err := crypto.ExtractECDSAPublicKeyFromFile(args.dataDirectory + "/" + wallet)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	commPrivKey, err := crypto.ExtractRSAKeyFromFile(args.dataDirectory + "/" + commitment)
	if err != nil {
		logger.Printf("%v\n", err)
		return err
	}

	if args.isRoot {
		multisigPrivKey, err := crypto.ExtractECDSAKeyFromFile(args.dataDirectory + "/" + multisig)
		if err != nil {
			logger.Printf("%v\n", err)
			return err
		}

		miner.Init(validatorPubKey, &multisigPrivKey.PublicKey, validatorPubKey, &commPrivKey.PublicKey, commPrivKey)
	} else {
		multisigPubKey, err := crypto.ExtractECDSAPublicKeyFromFile(args.dataDirectory + "/" + multisig)
		if err != nil {
			logger.Printf("%v\n", err)
			return err
		}

		rootPubKey, err := crypto.ExtractECDSAPublicKeyFromFile(args.dataDirectory + "/" + rootWallet)
		if err != nil {
			logger.Printf("%v\n", err)
			return err
		}

		rootCommPubKey, err := crypto.ExtractRSAPubKeyFromFile(args.dataDirectory + "/" + rootCommitment)
		if err != nil {
			logger.Printf("%v\n", err)
			return err
		}

		miner.Init(validatorPubKey, multisigPubKey, rootPubKey, rootCommPubKey, commPrivKey)
	}

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
	return nil
}

func (args startArgs) String() string {
	return fmt.Sprintf("Starting bazo miner with arguments \n"+
		"- My Address:\t\t\t %v\n"+
		"- Bootstrap Address:\t\t %v\n"+
		"- Data Directory:\t\t %v\n",
		args.myNodeAddress,
		args.bootstrapNodeAddress,
		args.dataDirectory)
}
