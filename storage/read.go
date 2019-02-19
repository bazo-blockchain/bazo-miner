package storage

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/boltdb/bolt"
)

//Always return nil if requested hash is not in the storage. This return value is then checked against by the caller
func ReadOpenBlock(hash [32]byte) (block *protocol.Block) {
	var encodedBlock []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(OPENBLOCKS_BUCKET))
		encodedBlock = b.Get(hash[:])
		return nil
	})

	if encodedBlock == nil {
		return nil
	}

	return block.Decode(encodedBlock)
}

func ReadOpenEpochBlock(hash [32]byte) (epochBlock *protocol.EpochBlock) {
	var encodedEpochBlock []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(OPENEPOCHBLOCK_BUCKET))
		encodedEpochBlock = b.Get(hash[:])
		return nil
	})

	if encodedEpochBlock == nil {
		return nil
	}

	return epochBlock.Decode(encodedEpochBlock)
}

func ReadClosedEpochBlock(hash [32]byte) (epochBlock *protocol.EpochBlock) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLOSEDEPOCHBLOCK_BUCKET))
		encodedBlock := b.Get(hash[:])
		epochBlock = epochBlock.Decode(encodedBlock)
		return nil
	})

	if epochBlock == nil {
		return nil
	}

	return epochBlock
}

func ReadClosedBlock(hash [32]byte) (block *protocol.Block) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLOSEDBLOCKS_BUCKET))
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
		b := tx.Bucket([]byte(LASTCLOSEDBLOCK_BUCKET))
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

func ReadLastClosedEpochBlock() (epochBlock *protocol.EpochBlock) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LASTCLOSEDEPOCHBLOCK_BUCKET))
		cb := b.Cursor()
		_, encodedBlock := cb.First()
		epochBlock = epochBlock.Decode(encodedBlock)
		return nil
	})

	if epochBlock == nil {
		return nil
	}

	return epochBlock
}

func ReadReceivedBlockStash() (receivedBlocks []*protocol.Block){
	return ReceivedBlockStash
}

func ReadINVALIDOpenTx(hash [32]byte) (transaction protocol.Transaction) {

	return txINVALIDMemPool[hash]
}

func ReadAllClosedBlocks() (allClosedBlocks []*protocol.Block) {
	if nextBlock := ReadLastClosedBlock(); nextBlock != nil {
		hasNext := true

		allClosedBlocks = append(allClosedBlocks, nextBlock)

		if nextBlock.Hash != [32]byte{} {
			for hasNext && nextBlock.Height > 1 {
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
	memPoolMutex.Lock()
	defer memPoolMutex.Unlock()
	return txMemPool[hash]
}


func GetMemPoolSize() int {
	memPoolMutex.Lock()
	defer memPoolMutex.Unlock()
	return len(txMemPool)
}

func ReadState() (state map[[64]byte]*protocol.Account){
	return State
}

//Needed for the miner to prepare a new block
func ReadAllOpenTxs() (allOpenTxs []protocol.Transaction) {
	memPoolMutex.Lock()
	defer memPoolMutex.Unlock()
	for key := range txMemPool {
		allOpenTxs = append(allOpenTxs, txMemPool[key])
	}
	return
}

//Personally I like it better to test (which tx type it is) here, and get returned the interface. Simplifies the code
func ReadClosedTx(hash [32]byte) (transaction protocol.Transaction) {
	if encodedTx := readClosedTx(CLOSEDFUNDS_BUCKET, hash); encodedTx != nil {
		var tx *protocol.FundsTx
		return tx.Decode(encodedTx)
	}

	if encodedTx := readClosedTx(CLOSEDACCS_BUCKET, hash); encodedTx != nil {
		var tx *protocol.ContractTx
		return tx.Decode(encodedTx)
	}

	if encodedTx := readClosedTx(CLOSEDCONFIGS_BUCKET, hash); encodedTx != nil {
		var tx *protocol.ConfigTx
		return tx.Decode(encodedTx)
	}

	if encodedTx := readClosedTx(CLOSEDSTAKES_BUCKET, hash); encodedTx != nil {
		var tx *protocol.StakeTx
		return tx.Decode(encodedTx)
	}

	return nil
}

func readClosedTx(bucketName string, hash [32]byte) (encodedTx []byte) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		encodedTx = b.Get(hash[:])
		return nil
	})
	return encodedTx
}

func ReadAccount(pubKey [64]byte) (acc *protocol.Account, err error) {
	if acc = State[pubKey]; acc != nil {
		return acc, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Acc (%x) not in the state.", pubKey[0:8]))
	}
}

func ReadRootAccount(pubKey [64]byte) (acc *protocol.Account, err error) {
	if IsRootKey(pubKey) {
		acc, err = ReadAccount(pubKey)
		return acc, err
	}

	return nil, err
}

func ReadGenesis() (genesis *protocol.Genesis, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(GENESIS_BUCKET))
		encoded := b.Get([]byte("genesis"))
		genesis = genesis.Decode(encoded)
		return nil
	})
	return genesis, err
}

func ReadFirstEpochBlock() (firstEpochBlock *protocol.EpochBlock, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLOSEDEPOCHBLOCK_BUCKET))
		encoded := b.Get([]byte("firstepochblock"))
		firstEpochBlock = firstEpochBlock.Decode(encoded)
		return nil
	})
	return firstEpochBlock, err
}

func ReadClosedState() (state *protocol.State, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(STATE_BUCKET))
		encoded := b.Get([]byte("state"))
		state = state.Decode(encoded)
		return nil
	})
	return state, err
}

func ReadValidatorMapping() (mapping *protocol.ValShardMapping, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(VALIDATOR_SHARD_MAPPING_BUCKET))
		encoded := b.Get([]byte("valshardmapping"))

		mapping = mapping.Decode(encoded)
		return nil
	})

	return mapping, err
}

func ReadReceivedBlockStashFromOtherShards() (*protocol.BlockStash){
	return ReceivedBlockStashFromOtherShards
}

func ReadBlockFromOwnStash(height int) *protocol.Block {
	for _, block := range OwnBlockStash {
		if(int(block.Height) == height){
			return block
		}
	}

	return nil
}

func ReadStateTransitionFromOwnStash(height int) *protocol.StateTransition {
	for _, st := range OwnStateTransitionStash {
		if(int(st.Height) == height){
			return st
		}
	}

	return nil
}