package miner

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"math/rand"
	"testing"
	"time"
)

func TestPrepareAndSortTxs(t *testing.T) {

	cleanAndPrepare()
	testsize := 100
	//fill the open storage with fundstx
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	for cnt := 0; cnt < testsize; cnt++ {
		accAHash := protocol.SerializeHashContent(accA.Address)
		accBHash := protocol.SerializeHashContent(accB.Address)
		tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, uint32(cnt), accAHash, accBHash, &PrivKeyA, &multiSignPrivKeyA)
		tx2, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, uint32(cnt), accBHash, accAHash, &PrivKeyB, &multiSignPrivKeyA)

		if verifyFundsTx(tx) {
			storage.WriteOpenTx(tx)
		}

		if verifyFundsTx(tx2) {
			storage.WriteOpenTx(tx2)
		}
	}

	//Add other tx types as well to make the test more challenging
	nullAddress := [64]byte{}
	for cnt := 0; cnt < testsize; cnt++ {
		tx, _, _ := protocol.ConstrAccTx(0x01, rand.Uint64()%100+1, nullAddress, &RootPrivKey)
		if verifyAccTx(tx) {
			storage.WriteOpenTx(tx)
		}
	}

	for cnt := 0; cnt < testsize; cnt++ {
		tx, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), uint8(rand.Uint32()%10+1), rand.Uint64()%2342873423, rand.Uint64()%1000+1, uint8(cnt), &RootPrivKey)

		//Don't mess with the minimum fee and block size
		if tx.Id == 3 || tx.Id == 1 {
			continue
		}
		if verifyConfigTx(tx) {
			storage.WriteOpenTx(tx)
		}
	}

	b := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	prepareBlock(b)
	finalizeBlock(b)

	//We could also use sort.IsSorted(...) bool, but manual check makes sure our sort interface is correct
	//this test ensures that all generated fundstx are included in the block, this is only possible if their
	//txcnt is sorted ascendingly
	if int(b.NrFundsTx) != testsize*2 {
		t.Errorf("NrFundsTx (%v) vs. testsize*2 (%v)\n", b.NrFundsTx, testsize*2)
	}
}
