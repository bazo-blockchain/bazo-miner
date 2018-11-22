package storage

import (
	"fmt"
	"log"
	"time"

	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/boltdb/bolt"
)

var (
	db                 *bolt.DB
	logger             *log.Logger
	State              = make(map[[64]byte]*protocol.Account)
	RootKeys           = make(map[[64]byte]*protocol.Account)
	txMemPool          = make(map[[32]byte]protocol.Transaction)
	AllClosedBlocksAsc []*protocol.Block
	BootstrapServer    string
	Buckets			   []string
)

const (
	ERROR_MSG 				= "Storage initialization aborted. Reason: "
	OPENBLOCKS_BUCKET 		= "openblocks"
	CLOSEDBLOCKS_BUCKET 	= "closedblocks"
	CLOSEDFUNDS_BUCKET 		= "closedfunds"
	CLOSEDACCS_BUCKET 		= "closedaccs"
	CLOSEDSTAKES_BUCKET 	= "closedstakes"
	CLOSEDCONFIGS_BUCKET	= "closedconfigs"
	LASTCLOSEDBLOCK_BUCKET 	= "lastclosedblock"
	ACCOUNTS_BUCKET			= "accounts"
)

//Entry function for the storage package
func Init(dbname string, bootstrapIpport string) error {
	BootstrapServer = bootstrapIpport
	logger = InitLogger()

	Buckets = []string {
		OPENBLOCKS_BUCKET,
		CLOSEDBLOCKS_BUCKET,
		CLOSEDFUNDS_BUCKET,
		CLOSEDACCS_BUCKET,
		CLOSEDSTAKES_BUCKET,
		CLOSEDCONFIGS_BUCKET,
		LASTCLOSEDBLOCK_BUCKET,
		ACCOUNTS_BUCKET,
	}

	var err error
	db, err = bolt.Open(dbname, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		logger.Fatal(ERROR_MSG, err)
		return err
	}

	for _, bucket := range Buckets {
		err = createBucket(bucket)
		if err != nil {
			return err
		}
	}

	err = loadAccountState()
	if err != nil {
		return err
	}

	return nil
}

func loadAccountState() error {
	accounts, err := ReadAccounts()
	if err != nil {
		return err
	}

	for _, acc := range accounts {
		State[acc.Address] = acc
	}
	return nil
}

func createBucket(bucketName string) (err error) {
	return db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte(bucketName))
		if err != nil {
			return fmt.Errorf(ERROR_MSG + " %s", err)
		}
		return nil
	})
}

func TearDown() {
	db.Close()
}
