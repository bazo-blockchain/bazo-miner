package main

import (
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"os"
)

func main() {
	logger := storage.InitLogger()

	if len(os.Args) != 7 {
		logger.Println("Usage: bazo-miner <dbname> <bootstrap ip:port> <this ip:port> <validatorfile> <multisigfile> <commitmentFile>")
		return
	}

	dbname := os.Args[1]
	bootstrapIpport := os.Args[2]
	thisIpport := os.Args[3]
	validatorFileName := os.Args[4]
	multisigFileName := os.Args[5]
	commFileName := os.Args[6]

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

	commPrivKey, err := protocol.ExtractRSAKeyFromFile(commFileName)
	if err != nil {
		logger.Printf("%v\n", err)
		return
	}

	miner.Init(&validatorPubKey, &multisigPubKey, &commPrivKey)
}
