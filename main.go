package main

import (
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"os"
)

func main() {
	logger := storage.InitLogger()

	//TODO Handle bootstrap address as app arg.
	if len(os.Args) != 7 {
		logger.Println("Usage: bazo-miner <dbname> <bootstrap ip:port> <this ip:port> <validatorfile> <seedfile> <multisigfile>")
		return
	}

	dbname := os.Args[1]
	bootstrapIpport := os.Args[2]
	thisIpport := os.Args[3]
	validatorFileName := os.Args[4]
	seedFileName := os.Args[5]
	multisigFileName := os.Args[6]

	storage.Init(dbname, bootstrapIpport)
	p2p.Init(thisIpport)

	validatorPubKey, _, err := storage.ExtractKeyFromFile(validatorFileName)
	if err != nil {
		logger.Printf("%v\n", err)
		return
	}

	multisigPubKey, _, err := storage.ExtractKeyFromFile(multisigFileName)
	if err != nil {
		logger.Printf("%v\n", err)
		return
	}

	miner.Init(&validatorPubKey, &multisigPubKey, seedFileName)
}
