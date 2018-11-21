package storage

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/boltdb/bolt"
)

func WriteOpenBlock(block *protocol.Block) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(OPENBLOCKS_BUCKET))
		return b.Put(block.Hash[:], block.Encode())
	})
}

func WriteClosedBlock(block *protocol.Block) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLOSEDBLOCKS_BUCKET))
		return b.Put(block.Hash[:], block.Encode())
	})
}

func WriteLastClosedBlock(block *protocol.Block) (err error) {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LASTCLOSEDBLOCK_BUCKET))
		return b.Put(block.Hash[:], block.Encode())
	})
}

//Changing the "tx" shortcut here and using "transaction" to distinguish between bolt's transactions
func WriteOpenTx(transaction protocol.Transaction) {
	txMemPool[transaction.Hash()] = transaction
}

func WriteClosedTx(transaction protocol.Transaction) error {
	var bucket string
	switch transaction.(type) {
	case *protocol.FundsTx:
		bucket = CLOSEDFUNDS_BUCKET
	case *protocol.AccTx:
		bucket = CLOSEDACCS_BUCKET
	case *protocol.ConfigTx:
		bucket = CLOSEDCONFIGS_BUCKET
	case *protocol.StakeTx:
		bucket = CLOSEDSTAKES_BUCKET
	}

	hash := transaction.Hash()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.Put(hash[:], transaction.Encode())
	})
}

func WriteAccount(account *protocol.Account) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ACCOUNTS_BUCKET))
		return b.Put(account.Address[:], account.Encode())
	})
}