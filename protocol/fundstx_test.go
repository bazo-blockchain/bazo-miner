package protocol

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestFundsTxSerialization(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	loopMax := int(rand.Uint32() % 10000)
	for i := 0; i < loopMax; i++ {
		tx, _ := ConstrFundsTx(0x01, rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accA.Address, accB.Address, PrivKeyA, nil)
		data := tx.Encode()
		var decodedTx *FundsTx
		decodedTx = decodedTx.Decode(data)

		//this is done by verify() which is outside protocol package, we're just testing serialization here
		decodedTx.From = accA.Address
		decodedTx.To = accB.Address

		if !reflect.DeepEqual(tx, decodedTx) {
			t.Errorf("FundsTx Serialization failed (%v) vs. (%v)\n", tx, decodedTx)
		}
	}
}
