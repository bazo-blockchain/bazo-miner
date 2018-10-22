package main

import (
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"os"
)

func main() {
	logger := storage.InitLogger()

	if len(os.Args) != 8 {
		logger.Println("Usage: bazo-miner <dbname> <bootstrap ip:port> <this ip:port> <validatorfile> <seedfile> <multisigfile> <commitmentfile>")
		return
	}

	dbname := os.Args[1]
	bootstrapIpport := os.Args[2]
	thisIpport := os.Args[3]
	validatorFileName := os.Args[4]
	seedFileName := os.Args[5]
	multisigFileName := os.Args[6]
	commFileName := os.Args[7]

	storage.Init(dbname, bootstrapIpport)
	p2p.Init(thisIpport)

	validatorPubKey, _, err := storage.ExtractECDSAKeyFromFile(validatorFileName)
	if err != nil {
		logger.Printf("%v\n", err)
		return
	}

	multisigPubKey, _, err := storage.ExtractECDSAKeyFromFile(multisigFileName)
	if err != nil {
		logger.Printf("%v\n", err)
		return
	}

	commPrivKey, err := storage.ExtractRSAKeyFromFile(commFileName)
	if err != nil {
		logger.Printf("%v\n", err)
		return
	}

	miner.Init(&validatorPubKey, &multisigPubKey, &commPrivKey, seedFileName)
}
