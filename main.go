package main

import (
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"os"
	"strings"
	"fmt"
	"strconv"
)

func main() {
	logger := storage.InitLogger()

	if len(os.Args) != 6 {
		logger.Println("Usage: bazo-miner <dbname> <ipport> <validator> <seedfile> <multisig>")
		return
	}

	if envInitRootBalance := os.Getenv("INITROOTBALANCE"); envInitRootBalance != "" {
		fmt.Printf("Using root balance from env %v\n", envInitRootBalance)
		parsedInitBalance, _  := strconv.ParseUint(envInitRootBalance, 0, 64)
		miner.InitialRootBalance = uint64(parsedInitBalance)
	}
	dbname := os.Args[1]
	ipport := os.Args[2]
	validator := os.Args[3]
	seedFileName := os.Args[4]
	multisig := os.Args[5]
	isBootstrap := strings.HasSuffix(ipport, storage.BOOTSTRAP_SERVER_PORT)

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
