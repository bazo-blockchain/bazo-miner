package main

import (
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"os"
	"strings"
)

func main() {
	logger := storage.InitLogger()

	if len(os.Args) != 6 {
		logger.Println("Usage: bazo-miner <dbname> <ipport> <validator> <seedfile> <multisig>")
		return
	}

	dbname := os.Args[1]
	ipport := os.Args[2]
	validator := os.Args[3]
	seedFileName := os.Args[4]
	multisig := os.Args[5]
	isBootstrap := false

	if strings.HasSuffix(os.Args[2], storage.BOOTSTRAP_SERVER_PORT){
		isBootstrap = true
	}

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

	miner.Init(&validatorPubKey, &multisigPubKey, seedFileName, isBootstrap)
}
