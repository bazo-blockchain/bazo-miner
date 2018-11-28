package miner

import (
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"math/rand"
	"testing"
	"time"

	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

func TestPrepareAndSortTxs(t *testing.T) {
	cleanAndPrepare()

	testsize := 100
	//fill the open storage with fundstx
	randVar := rand.New(rand.NewSource(time.Now().Unix()))
	for cnt := 0; cnt < testsize; cnt++ {
		tx, _ := protocol.ConstrFundsTx(0x01, randVar.Uint64()%100+1, randVar.Uint64()%100+1, uint32(cnt), accA.Address, accB.Address, PrivKeyAccA, nil)
		tx2, _ := protocol.ConstrFundsTx(0x01, randVar.Uint64()%100+1, randVar.Uint64()%100+1, uint32(cnt), accB.Address, accA.Address, PrivKeyAccB, nil)

		if verifyFundsTx(tx) {
			storage.WriteOpenTx(tx)
		}

		if verifyFundsTx(tx2) {
			storage.WriteOpenTx(tx2)
		}
	}

	//Add other tx types as well to make the test more challenging
	for cnt := 0; cnt < testsize; cnt++ {
		tx, _, _ := protocol.ConstrContractTx(0x01, randVar.Uint64()%100+1, PrivKeyRoot, nil, nil)
		if verifyContractTx(tx) {
			storage.WriteOpenTx(tx)
		}
	}

	for cnt := 0; cnt < testsize; cnt++ {
		tx, _ := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), uint8(randVar.Uint32()%10+1), randVar.Uint64()%2342873423, randVar.Uint64()%1000+1, uint8(cnt), PrivKeyRoot)

		//Don't mess with the minimum fee and block size
		if tx.Id == 3 || tx.Id == 1 {
			continue
		}
		if verifyConfigTx(tx) {
			storage.WriteOpenTx(tx)
		}
	}

	b := newBlock([32]byte{}, [crypto.COMM_PROOF_LENGTH]byte{}, 1)
	prepareBlock(b)
	finalizeBlock(b)

	//We could also use sort.IsSorted(...) bool, but manual check makes sure our sort interface is correct
	//this test ensures that all generated fundstx are included in the block, this is only possible if their
	//txcnt is sorted ascendingly
	if int(b.NrFundsTx) != testsize*2 {
		t.Errorf("NrFundsTx (%v) vs. testsize*2 (%v)\n", b.NrFundsTx, testsize*2)
	}
}
