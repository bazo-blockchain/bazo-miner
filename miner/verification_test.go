package miner

import (
	"github.com/sfontanach/bazo-miner/protocol"
	"math/rand"
	"testing"
	"time"
)

func TestFundsTxVerification(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	loopMax := int(rand.Uint64() % 1000)
	accAHash := protocol.SerializeHashContent(accA.Address)
	accBHash := protocol.SerializeHashContent(accB.Address)
	for i := 0; i < loopMax; i++ {
		tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accAHash, accBHash, &PrivKeyA)
		if verifyFundsTx(tx) == false {
			t.Errorf("Tx could not be verified: \n%v", tx)
		}
	}
}

func TestAccTx(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	//Creating some root-signed new accounts
	loopMax := int(rand.Uint64() % 1000)
	for i := 0; i <= loopMax; i++ {
		tx, _, _ := protocol.ConstrAccTx(0, rand.Uint64()%100+1, &RootPrivKey)
		if verifyAccTx(tx) == false {
			t.Errorf("AccTx could not be verified: %v\n", tx)
		}
	}
}

func TestConfigTx(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	//creating some root-signed config txs
	tx, err := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 1, 5000, rand.Uint64(), 0, &RootPrivKey)
	tx2, err2 := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 2, 5000, rand.Uint64(), 0, &RootPrivKey)
	tx3, err3 := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 3, 5000, rand.Uint64(), 0, &RootPrivKey)
	tx4, err4 := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 4, 5000, rand.Uint64(), 0, &RootPrivKey)
	tx5, err5 := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 5, 5000, rand.Uint64(), 0, &RootPrivKey)

	//Add an invalid configTx, should not be accepted
	txfail, err6 := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 20, 5000, rand.Uint64(), 0, &RootPrivKey)

	if (verifyConfigTx(tx) == false || err != nil) &&
		(verifyConfigTx(tx2) == false || err2 != nil) &&
		(verifyConfigTx(tx3) == false || err3 != nil) &&
		(verifyConfigTx(tx4) == false || err4 != nil) &&
		(verifyConfigTx(tx5) == false || err5 != nil) &&
		(verifyConfigTx(txfail) == true || err6 != nil) {
		t.Error("ConfigTx verification malfunctioning!")
	}
}
