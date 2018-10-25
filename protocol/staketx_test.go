package protocol

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestStakeTxSerialization(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	accAHash := SerializeHashContent(accA.Address)
	loopMax := int(rand.Uint32() % 10000)
	for i := 0; i < loopMax; i++ {
		fee := rand.Uint64()%10 + 1
		isStaking := rand.Intn(2) != 0

		tx, _ := ConstrStakeTx(0x01, fee, isStaking, accAHash, &PrivKeyA, &CommKeyA)
		data := tx.Encode()
		var decodedTx *StakeTx
		decodedTx = decodedTx.Decode(data)

		//this is done by verify() which is outside protocol package, we're just testing serialization here
		//decodedTx.Fee = fee
		//decodedTx.Account = accAHash
		//decodedTx.IsStaking = false
		//decodedTx.HashedSeed = hashedSeed

		if !reflect.DeepEqual(tx.Hash(), decodedTx.Hash()) {
			t.Errorf("StakeTx Serialization failed (%v) vs. (%v)\n", tx, decodedTx)
		}

		if !reflect.DeepEqual(tx, decodedTx) {
			t.Errorf("StakeTx Serialization failed (%v) vs. (%v)\n", tx, decodedTx)
		}
	}
}
