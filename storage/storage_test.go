package storage

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"math/rand"
	"testing"
	"time"
)

//In-memory, k/v storage is tested with the test below
func TestReadWriteDeleteTx(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))

	accAHash := protocol.SerializeHashContent(accA.Address)
	accBHash := protocol.SerializeHashContent(accB.Address)

	var hashFundsSlice []*protocol.FundsTx
	var hashAccSlice []*protocol.AccTx
	var hashConfigSlice []*protocol.ConfigTx

	testsize := 1000

	loopMax := testsize
	for i := 0; i < loopMax; i++ {
		tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accAHash, accBHash, &PrivKeyA, nil)
		WriteOpenTx(tx)
		hashFundsSlice = append(hashFundsSlice, tx)
	}

	loopMax = testsize
	for i := 0; i < loopMax; i++ {
		tx, _, _ := protocol.ConstrAccTx(0, rand.Uint64()%100+1, accA.Address, &RootPrivKey)
		tx.Hash()
		WriteOpenTx(tx)
		hashAccSlice = append(hashAccSlice, tx)
	}

	//Restricted to 256, because the number of configTxs is stored in a uint8 in blocks
	loopMax = 256
	for cnt := 0; cnt < loopMax; cnt++ {
		tx, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), uint8(rand.Uint32()%5+1), rand.Uint64()%2342873423, rand.Uint64()%1000+1, uint8(cnt), &RootPrivKey)
		hashConfigSlice = append(hashConfigSlice, tx)
		WriteOpenTx(tx)
	}

	for _, tx := range hashFundsSlice {
		if ReadOpenTx(tx.Hash()) == nil {
			t.Errorf("Error writing transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashAccSlice {
		if ReadOpenTx(tx.Hash()) == nil {
			t.Errorf("Error writing transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashConfigSlice {
		if ReadOpenTx(tx.Hash()) == nil {
			t.Errorf("Error writing transaction hash: %x\n", tx)
		}
	}

	//Read all open txs, received in random order
	opentxs := ReadAllOpenTxs()

	//Comparing the total number of txs should be enough
	totaltxs := len(hashConfigSlice)+len(hashFundsSlice)+len(hashAccSlice)
	if len(opentxs) != totaltxs {
		t.Error("ReadAllOpenTxs() returned an invalid list of transactions\n")
	}

	//Deleting open txs
	for _, tx := range hashFundsSlice {
		DeleteOpenTx(tx)
	}

	for _, tx := range hashAccSlice {
		DeleteOpenTx(tx)
	}

	for _, tx := range hashConfigSlice {
		DeleteOpenTx(tx)
	}

	//Make sure all txs are actually deleted
	for _, tx := range hashFundsSlice {
		if ReadOpenTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashAccSlice {
		if ReadOpenTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashConfigSlice {
		if ReadOpenTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}

	//Same with k/v-based closed tx storage
	for _, tx := range hashAccSlice {
		WriteClosedTx(tx)
	}

	for _, tx := range hashFundsSlice {
		WriteClosedTx(tx)
	}

	for _, tx := range hashConfigSlice {
		WriteClosedTx(tx)
	}

	for _, tx := range hashAccSlice {
		if ReadClosedTx(tx.Hash()) == nil {
			t.Errorf("Error writing to k/v storage: %x\n", tx)
		}
	}

	for _, tx := range hashFundsSlice {
		if ReadClosedTx(tx.Hash()) == nil {
			t.Errorf("Error writing to k/v storage: %x\n", tx)
		}
	}

	for _, tx := range hashConfigSlice {
		if ReadClosedTx(tx.Hash()) == nil {
			t.Errorf("Error writing to k/v storage: %x\n", tx)
		}
	}

	//Delete transactions from closed storage
	for _, tx := range hashAccSlice {
		DeleteClosedTx(tx)
	}

	for _, tx := range hashFundsSlice {
		DeleteClosedTx(tx)
	}

	for _, tx := range hashConfigSlice {
		DeleteClosedTx(tx)
	}

	//Make sure all txs are actually deleted
	for _, tx := range hashAccSlice {
		if ReadClosedTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashFundsSlice {
		if ReadClosedTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashConfigSlice {
		if ReadClosedTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}
}

//Disk-based k/v storage
func TestReadWriteDeleteBlock(t *testing.T) {

	//No panic
	DeleteOpenBlock([32]byte{'0'})

	b, b2, b3 := new(protocol.Block), new(protocol.Block), new(protocol.Block)
	b.Hash = [32]byte{'0'}
	b2.Hash = [32]byte{'1'}
	b3.Hash = [32]byte{'2'}
	WriteOpenBlock(b)
	WriteOpenBlock(b2)
	WriteOpenBlock(b3)

	if ReadOpenBlock(b.Hash) == nil || ReadOpenBlock(b2.Hash) == nil || ReadOpenBlock(b3.Hash) == nil {
		t.Error("Failed to write block to open block storage.\n")
	}

	newb1 := ReadOpenBlock(b.Hash)
	newb2 := ReadOpenBlock(b2.Hash)
	newb3 := ReadOpenBlock(b3.Hash)

	DeleteOpenBlock(newb1.Hash)
	DeleteOpenBlock(newb2.Hash)
	DeleteOpenBlock(newb3.Hash)

	WriteClosedBlock(newb1)
	WriteClosedBlock(newb2)
	WriteClosedBlock(newb3)

	if ReadOpenBlock(newb1.Hash) != nil ||
		ReadOpenBlock(newb2.Hash) != nil ||
		ReadOpenBlock(newb3.Hash) != nil ||
		ReadClosedBlock(b.Hash) == nil ||
		ReadClosedBlock(b2.Hash) == nil ||
		ReadClosedBlock(b3.Hash) == nil {
		t.Error("Failed to write block to kv storage.\n")
	}

	DeleteClosedBlock(newb1.Hash)
	DeleteClosedBlock(newb2.Hash)
	DeleteClosedBlock(newb3.Hash)

	if ReadClosedBlock(b.Hash) != nil ||
		ReadClosedBlock(b2.Hash) != nil ||
		ReadClosedBlock(b3.Hash) != nil {
		t.Error("Failed to delete block from kv storage.\n")
	}
}
