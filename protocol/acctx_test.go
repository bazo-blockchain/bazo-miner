package protocol

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestAccTx_Serialization(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	nullAddress := [64]byte{}
	loopMax := int(rand.Uint32() % 10000)
	for i := 1; i < loopMax; i++ {
		tx, _, _ := ConstrAccTx(0, rand.Uint64()%100+1, nullAddress, &RootPrivKey)
		data := tx.Encode()
		var decodedTx *AccTx
		decodedTx = decodedTx.Decode(data)
		if !reflect.DeepEqual(tx, decodedTx) {
			t.Errorf("AccTx Serialization failed (%v) vs. (%v)\n", tx, decodedTx)
		}
	}
}

func TestAccTx_Serialization_Contract(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	nullAddress := [64]byte{}
	loopMax := int(rand.Uint32() % 10000)
	for i := 1; i < loopMax; i++ {
		tx, _, _ := ConstrAccTx(0, rand.Uint64()%100+1, nullAddress, &RootPrivKey, RandomBytes(), []ByteArray{[]byte{1}})
		data := tx.Encode()
		var decodedTx *AccTx
		decodedTx = decodedTx.Decode(data)
		if !reflect.DeepEqual(tx, decodedTx) {
			t.Errorf("ContractTx Serialization failed (%v) vs. (%v)\n", tx, decodedTx)
		}
	}
}
