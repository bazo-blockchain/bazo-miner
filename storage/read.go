package storage

import (
	"../protocol"
	"github.com/boltdb/bolt"
)

//Always return nil if requested hash is not in the storage. This return value is then checked against by the caller
func ReadOpenBlock(hash [32]byte) (block *protocol.Block) {

	var encodedBlock []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("openblocks"))
		encodedBlock = b.Get(hash[:])
		return nil
	})

	if encodedBlock == nil {
		return nil
	}

	return block.Decode(encodedBlock)
}

func ReadClosedBlock(hash [32]byte) (block *protocol.Block) {

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedblocks"))
		encodedBlock := b.Get(hash[:])
		block = block.Decode(encodedBlock)
		return nil
	})

	if block == nil {
		return nil
	}

	return block
}

func ReadLastClosedBlock() (block *protocol.Block) {

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("lastclosedblock"))
		cb := b.Cursor()
		_, encodedBlock := cb.First()
		block = block.Decode(encodedBlock)
		return nil
	})

	if block == nil {
		return nil
	}

	return block
}

func ReadAllClosedBlocks() (allClosedBlocks []*protocol.Block) {
	if nextBlock := ReadLastClosedBlock(); nextBlock != nil {
		hasNext := true

		allClosedBlocks = append(allClosedBlocks, nextBlock)

		if nextBlock.Hash != [32]byte{} {
			for hasNext {
				nextBlock = ReadClosedBlock(nextBlock.PrevHash)
				allClosedBlocks = append(allClosedBlocks, nextBlock)
				if nextBlock.Hash == [32]byte{} {
					hasNext = false
				}
			}
		}
	}

	return allClosedBlocks
}

func ReadOpenTx(hash [32]byte) (transaction protocol.Transaction) {

	return txMemPool[hash]
}

//Needed for the miner to prepare a new block
func ReadAllOpenTxs() (allOpenTxs []protocol.Transaction) {

	for key := range txMemPool {
		allOpenTxs = append(allOpenTxs, txMemPool[key])
	}
	return
}

//Personally I like it better to test (which tx type it is) here, and get returned the interface. Simplifies the code
func ReadClosedTx(hash [32]byte) (transaction protocol.Transaction) {

	var encodedTx []byte
	var fundstx *protocol.FundsTx
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedfunds"))
		encodedTx = b.Get(hash[:])
		return nil
	})
	if encodedTx != nil {
		return fundstx.Decode(encodedTx)
	}

	var acctx *protocol.AccTx
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedaccs"))
		encodedTx = b.Get(hash[:])
		return nil
	})
	if encodedTx != nil {
		return acctx.Decode(encodedTx)
	}

	var configtx *protocol.ConfigTx
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedconfigs"))
		encodedTx = b.Get(hash[:])
		return nil
	})
	if encodedTx != nil {
		return configtx.Decode(encodedTx)
	}

	var staketx *protocol.StakeTx
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("closedstakes"))
		encodedTx = b.Get(hash[:])
		return nil
	})
	if encodedTx != nil {
		return staketx.Decode(encodedTx)
	}
	return nil
}
