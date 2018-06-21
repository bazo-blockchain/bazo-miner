package miner

import (
	"github.com/bazo-blockchain/bazo-miner/storage"
	"testing"
)

//Tests cases 1) when new block is received that belongs to a longer chain and 2) when new block is received
//that is shorter than the current chain
func TestGetBlockSequences(t *testing.T) {

	cleanAndPrepare()

	b := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	createBlockWithTxs(b)
	finalizeBlock(b)
	validate(b)

	b2 := newBlock(b.Hash, [32]byte{}, [32]byte{}, b.Height+1)
	createBlockWithTxs(b2)
	finalizeBlock(b2)
	validate(b2)

	b3 := newBlock(b2.Hash, [32]byte{}, [32]byte{}, b2.Height+1)
	createBlockWithTxs(b3)
	finalizeBlock(b3)

	rollback, validate := getBlockSequences(b3)

	if len(rollback) != 0 {
		t.Error("Rollback shouldn't execute here\n")
	}

	if len(validate) != 1 || validate[0].Hash != b3.Hash {
		t.Error("Wrong validation sequence\n")
	}

	//PoW needs lastBlock, have to set it manually
	lastBlock = storage.ReadClosedBlock([32]byte{})
	c := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	createBlockWithTxs(c)
	finalizeBlock(c)
	storage.WriteOpenBlock(c)

	//PoW needs lastBlock, have to set it manually
	lastBlock = c
	c2 := newBlock(c.Hash, [32]byte{}, [32]byte{}, c.Height+1)
	createBlockWithTxs(c2)
	finalizeBlock(c2)
	storage.WriteOpenBlock(c2)

	//PoW needs lastBlock, have to set it manually
	lastBlock = c2
	c3 := newBlock(c2.Hash, [32]byte{}, [32]byte{}, c.Height+1)
	createBlockWithTxs(c3)
	finalizeBlock(c3)

	lastBlock = b2
	//Blockchain now: genesis <- b <- b2
	//New Blockchain of longer size: genesis <- c <- c2 <- c3
	rollback, validate = getBlockSequences(c3)

	//Rollback slice needs to include b2 and b (in that order)
	if len(rollback) != 2 ||
		rollback[0].Hash != b2.Hash ||
		rollback[1].Hash != b.Hash {
		t.Error("Rollback of current chain failed\n")
	}

	if len(validate) != 3 ||
		validate[0].Hash != c.Hash ||
		validate[1].Hash != c2.Hash ||
		validate[2].Hash != c3.Hash {
		t.Error("Validation failed\n")
	}

	cleanAndPrepare()
	//Make sure that another chain of equal length does not get activated
	b = newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	createBlockWithTxs(b)
	finalizeBlock(b)
	validate(b)

	b2 = newBlock(b.Hash, [32]byte{}, [32]byte{}, b.Height+1)
	createBlockWithTxs(b2)
	finalizeBlock(b2)
	validate(b2)

	b3 = newBlock(b2.Hash, [32]byte{}, [32]byte{}, b2.Height+1)
	createBlockWithTxs(b3)
	finalizeBlock(b3)
	validate(b3)

	//Blockchain now: genesis <- b <- b2 <- b3
	//Competing chain: genesis <- c <- c2 <- c3
	lastBlock = storage.ReadClosedBlock([32]byte{})
	c = newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	createBlockWithTxs(c)
	finalizeBlock(c)
	storage.WriteOpenBlock(c)

	lastBlock = c
	c2 = newBlock(c.Hash, [32]byte{}, [32]byte{}, c.Height+1)
	createBlockWithTxs(c2)
	finalizeBlock(c2)
	storage.WriteOpenBlock(c2)

	lastBlock = c2
	c3 = newBlock(c2.Hash, [32]byte{}, [32]byte{}, c2.Height+1)
	createBlockWithTxs(c3)
	finalizeBlock(c3)

	//Make sure that the new blockchain of equal length does not get activated
	lastBlock = b3
	rollback, validate = getBlockSequences(c3)
	if rollback != nil || validate != nil {
		t.Error("Did not properly detect longest chain\n")
	}
}

//Test whether we get the new proper chain (we leverage the fact that open storage is checked so we don't need
//to need network functionality for that test
func TestGetNewChain(t *testing.T) {

	cleanAndPrepare()
	b := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	createBlockWithTxs(b)
	finalizeBlock(b)
	validate(b)

	b2 := newBlock(b.Hash, [32]byte{}, [32]byte{}, b.Height+1)
	createBlockWithTxs(b2)
	finalizeBlock(b2)

	ancestor, newChain := getNewChain(b2)

	if ancestor.Hash != b.Hash {
		t.Errorf("Hash mismatch: %x vs. %x\n", ancestor.Hash, b.Hash)
	}
	if len(newChain) != 1 || newChain[0].Hash != b2.Hash {
		t.Error("Wrong new chain\n")
	}

	//Blockchain now: genesis <- b
	//New chain: genesis <- c <- c2
	lastBlock = storage.ReadClosedBlock([32]byte{})
	c := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	createBlockWithTxs(c)
	finalizeBlock(c)
	storage.WriteOpenBlock(c)

	lastBlock = c
	c2 := newBlock(c.Hash, [32]byte{}, [32]byte{}, c.Height+1)
	createBlockWithTxs(c2)
	finalizeBlock(c2)

	lastBlock = b
	ancestor, newChain = getNewChain(c2)

	if ancestor.Hash != [32]byte{} {
		t.Errorf("Hash mismatch")
	}

	if len(newChain) != 2 || newChain[0].Hash != c.Hash || newChain[1].Hash != c2.Hash {
		t.Error("Wrong new chain\n")
	}
}
