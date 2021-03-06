package storage

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/boltdb/bolt"
)

func WriteOpenBlock(block *protocol.Block) (err error) {

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openblocks"))
		err := b.Put(block.Hash[:], block.Encode())
		return err
	})

	return err
}

func WriteClosedBlock(block *protocol.Block) (err error) {

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedblocks"))
		err := b.Put(block.Hash[:], block.Encode())
		return err
	})

	return err
}

func WriteLastClosedBlock(block *protocol.Block) (err error) {

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("lastclosedblock"))
		err := b.Put(block.Hash[:], block.Encode())
		return err
	})

	return err
}

//Changing the "tx" shortcut here and using "transaction" to distinguish between bolt's transactions
func WriteOpenTx(transaction protocol.Transaction) {

	txMemPool[transaction.Hash()] = transaction
}

func WriteINVALIDOpenTx(transaction protocol.Transaction) {

	txINVALIDMemPool[transaction.Hash()] = transaction
}
func WriteToReceivedStash(block *protocol.Block) {

	//Only write it to stash if it is not in there already.
	if !blockAlreadyInStash(receivedBlockStash, block.Hash) {
		receivedBlockStash = append(receivedBlockStash, block)

		//When lenght of stash is > 50 --> Remove first added Block
		if len(receivedBlockStash) > 50 {
			receivedBlockStash = append(receivedBlockStash[:0], receivedBlockStash[1:]...)
		}
	}
}

func blockAlreadyInStash(slice []*protocol.Block, newBlockHash [32]byte) bool {
	for _, blockInStash := range slice {
		if blockInStash.Hash == newBlockHash {
			return true
		}
	}
	return false
}

func WriteClosedTx(transaction protocol.Transaction) (err error) {

	var bucket string
	switch transaction.(type) {
	case *protocol.FundsTx:
		bucket = "closedfunds"
	case *protocol.AccTx:
		bucket = "closedaccs"
	case *protocol.ConfigTx:
		bucket = "closedconfigs"
	case *protocol.StakeTx:
		bucket = "closedstakes"
	}

	hash := transaction.Hash()
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put(hash[:], transaction.Encode())
		return err
	})

	PrintMemPoolSize()
	return err
}
