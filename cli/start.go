package cli

import (
	"fmt"
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
}

func GetStartCommand() cli.Command {
	return cli.Command{
		Name:  "start",
		Usage: "start the miner",
		Action: func(c *cli.Context) error {
			args := &startArgs{
				dataDirectory:        	c.String("dataDir"),
				myNodeAddress:        	c.String("address"),
				bootstrapNodeAddress: 	c.String("bootstrap"),
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

			return Start(args)
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
				Name:  "confirm",
				Usage: "User must press enter before starting the miner",
			},
		},
	}
}

func Start(args *startArgs) error {
	var firstStart = false
	if _, err := os.Stat(args.dataDirectory); os.IsNotExist(err) {
		err = os.MkdirAll(args.dataDirectory, 0755)
		if err != nil {
			return err
		}
		firstStart = true
	}

	const (
		database       = "store.db"
		wallet         = "wallet.key"
		commitment     = "commitment.key"
	)

	storage.Init(args.dataDirectory+"/" + database, args.bootstrapNodeAddress)
	p2p.Init(args.myNodeAddress)

	validatorPubKey, err := crypto.ExtractECDSAPublicKeyFromFile(args.dataDirectory + "/" + wallet)
	if err != nil {
		return err
	}

	commPrivKey, err := crypto.ExtractRSAKeyFromFile(args.dataDirectory + "/" + commitment)
	if err != nil {
		return err
	}

	// Check if executor is root and if it's the first start
	if p2p.IsBootstrap() && firstStart {
		return miner.InitFirstStart(validatorPubKey, commPrivKey)
	} else {
		return miner.Init(validatorPubKey, commPrivKey)
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
