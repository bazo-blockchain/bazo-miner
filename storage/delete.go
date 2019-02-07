package storage

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/boltdb/bolt"
)

//There exist open/closed buckets and closed tx buckets for all types (open txs are in volatile storage)
func DeleteOpenBlock(hash [32]byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(OPENBLOCKS_BUCKET))
		return b.Delete(hash[:])
	})
}

func DeleteClosedBlock(hash [32]byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLOSEDBLOCKS_BUCKET))
		return b.Delete(hash[:])
	})
}

func DeleteClosedEpochBlock(hash [32]byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLOSEDEPOCHBLOCK_BUCKET))
		return b.Delete(hash[:])
	})
}

func DeleteState() error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(STATE_BUCKET))
		return b.Delete([]byte("state"))
	})
}

func DeleteOpenEpochBlock(hash [32]byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(OPENEPOCHBLOCK_BUCKET))
		return b.Delete(hash[:])
	})
}

func DeleteINVALIDOpenTx(transaction protocol.Transaction) {
	delete(txINVALIDMemPool, transaction.Hash())
}

func DeleteLastClosedBlock(hash [32]byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LASTCLOSEDBLOCK_BUCKET))
		return b.Delete(hash[:])
	})
}

func DeleteAllLastClosedBlock() error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LASTCLOSEDBLOCK_BUCKET))
		return b.ForEach(func(k, v []byte) error {
			return b.Delete(k)
		})
	})
}

func DeleteAllLastClosedEpochBlock() error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LASTCLOSEDEPOCHBLOCK_BUCKET))
		return b.ForEach(func(k, v []byte) error {
			return b.Delete(k)
		})
	})
}

func DeleteOpenTx(transaction protocol.Transaction) {
	memPoolMutex.Lock()
	defer memPoolMutex.Unlock()
	delete(txMemPool, transaction.Hash())
}

func DeleteClosedTx(transaction protocol.Transaction) error {
	var bucket string
	switch transaction.(type) {
	case *protocol.FundsTx:
		bucket = CLOSEDFUNDS_BUCKET
	case *protocol.ContractTx:
		bucket = CLOSEDACCS_BUCKET
	case *protocol.ConfigTx:
		bucket = CLOSEDCONFIGS_BUCKET
	case *protocol.StakeTx:
		bucket = CLOSEDSTAKES_BUCKET
	}

	hash := transaction.Hash()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.Delete(hash[:])
	})
}

func DeleteAccount(address [64]byte) {
	delete(State, address)
}

func DeleteAll() (err error) {
	//Delete in-memory storage
	for key := range txMemPool {
		delete(txMemPool, key)
	}

	//Delete disk-based storage
	for _, bucket := range Buckets {
		err = clearBucket(bucket)
		if err != nil {
			return err
		}
	}

	return nil
}

func clearBucket(bucketName string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.ForEach(func(k, v []byte) error {
			return b.Delete(k)
		})
	})
}