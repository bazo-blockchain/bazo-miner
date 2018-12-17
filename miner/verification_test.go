package miner

import (
	"math/rand"
	"testing"
	"time"

	"github.com/bazo-blockchain/bazo-miner/protocol"
)

func TestFundsTxVerification(t *testing.T) {
	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	loopMax := int(randVar.Uint64() % 1000)
	for i := 0; i < loopMax; i++ {
		tx, _ := protocol.NewSignedFundsTx(0x01, randVar.Uint64()%100000+1, randVar.Uint64()%10+1, uint32(i), accA.Address, accB.Address, PrivKeyAccA, nil)
		if verifyFundsTx(tx) != nil {
			t.Errorf("Tx could not be verified: \n%v", tx)
		}
	}
}

func TestContractTx(t *testing.T) {
	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	//Creating some root-signed new accounts
	loopMax := int(randVar.Uint64() % 1000)
	for i := 0; i <= loopMax; i++ {
		tx, _, _ := protocol.ConstrContractTx(0, randVar.Uint64()%100+1, PrivKeyRoot, nil, nil)
		if verifyContractTx(tx) != nil {
			t.Errorf("ContractTx could not be verified: %v\n", tx)
		}
	}
}

func TestConfigTx(t *testing.T) {
	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	//creating some root-signed config txs
	tx, err := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 1, 5000, randVar.Uint64(), 0, PrivKeyRoot)
	tx2, err2 := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 2, 5000, randVar.Uint64(), 0, PrivKeyRoot)
	tx3, err3 := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 3, 5000, randVar.Uint64(), 0, PrivKeyRoot)
	tx4, err4 := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 4, 5000, randVar.Uint64(), 0, PrivKeyRoot)
	tx5, err5 := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 5, 5000, randVar.Uint64(), 0, PrivKeyRoot)

	//Add an invalid configTx, should not be accepted
	txfail, err6 := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 20, 5000, randVar.Uint64(), 0, PrivKeyRoot)

	if (verifyConfigTx(tx) != nil || err != nil) &&
		(verifyConfigTx(tx2) != nil || err2 != nil) &&
		(verifyConfigTx(tx3) != nil || err3 != nil) &&
		(verifyConfigTx(tx4) != nil || err4 != nil) &&
		(verifyConfigTx(tx5) != nil || err5 != nil) &&
		(verifyConfigTx(txfail) != nil || err6 != nil) {
		t.Error("ConfigTx verification malfunctioning!")
	}
}
