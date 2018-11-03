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
	accAHash := crypto.SerializeHashContent(accA.Address)
	accBHash := crypto.SerializeHashContent(accB.Address)
	for i := 0; i < loopMax; i++ {
		tx, _ := protocol.ConstrFundsTx(0x01, randVar.Uint64()%100000+1, randVar.Uint64()%10+1, uint32(i), accAHash, accBHash, &PrivKeyAccA, &PrivKeyMultiSig, nil)
		if verifyFundsTx(tx) == false {
			t.Errorf("Tx could not be verified: \n%v", tx)
		}
	}
}

func TestAccTx(t *testing.T) {
	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	//Creating some root-signed new accounts
	nullAccount := [64]byte{1}
	loopMax := int(randVar.Uint64() % 1000)
	for i := 0; i <= loopMax; i++ {
		tx, _, _ := protocol.ConstrAccTx(0, randVar.Uint64()%100+1, nullAccount, &PrivKeyRoot, nil, nil)
		if verifyAccTx(tx) == false {
			t.Errorf("AccTx could not be verified: %v\n", tx)
		}
	}
}

func TestConfigTx(t *testing.T) {
	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	//creating some root-signed config txs
	tx, err := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 1, 5000, randVar.Uint64(), 0, &PrivKeyRoot)
	tx2, err2 := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 2, 5000, randVar.Uint64(), 0, &PrivKeyRoot)
	tx3, err3 := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 3, 5000, randVar.Uint64(), 0, &PrivKeyRoot)
	tx4, err4 := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 4, 5000, randVar.Uint64(), 0, &PrivKeyRoot)
	tx5, err5 := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 5, 5000, randVar.Uint64(), 0, &PrivKeyRoot)

	//Add an invalid configTx, should not be accepted
	txfail, err6 := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 20, 5000, randVar.Uint64(), 0, &PrivKeyRoot)

	if (verifyConfigTx(tx) == false || err != nil) &&
		(verifyConfigTx(tx2) == false || err2 != nil) &&
		(verifyConfigTx(tx3) == false || err3 != nil) &&
		(verifyConfigTx(tx4) == false || err4 != nil) &&
		(verifyConfigTx(tx5) == false || err5 != nil) &&
		(verifyConfigTx(txfail) == true || err6 != nil) {
		t.Error("ConfigTx verification malfunctioning!")
	}
}
