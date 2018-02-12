package main

import (
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"os"
	"fmt"
)

func main() {
	var localConn, dbname, validatorAccFile string

	dbname = os.Args[1]
	localConn = os.Args[2]

	storage.Init(localConn, dbname)
	p2p.Init(localConn)

	//Validate from existing account possible by submitting the file name of the key file.
	//Otherwise, a root account will be initialized.
	if len(os.Args) == 4 {
		if _, err := os.Stat(os.Args[3]); os.IsNotExist(err) {
			fmt.Printf("%s: file does not exist", os.Args[3])
			return
		}
		validatorAccFile = os.Args[3]
	}
	miner.Init(validatorAccFile)
}
