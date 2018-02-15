package main

import (
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"os"
)

func main() {

	dbname := os.Args[1]
	ipport := os.Args[2]
	validator := os.Args[3]
	multisig := os.Args[4]

	storage.Init(dbname, ipport)
	p2p.Init(ipport)

	validatorPubKey, _, err := storage.ExtractKeyFromFile(validator)
	if err != nil {
		logger.Printf("%v\n", err)
		return
	}

	multisigPubKey, _, err := storage.ExtractKeyFromFile(multisig)
	if err != nil {
		logger.Printf("%v\n", err)
		return
	}

	miner.Init(&validatorPubKey, &multisigPubKey)
}
