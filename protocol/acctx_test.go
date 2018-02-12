package protocol

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestAccTxSerialization(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	loopMax := int(rand.Uint32() % 10000)
	for i := 1; i < loopMax; i++ {
		tx, _, _ := ConstrAccTx(0, rand.Uint64()%100+1, &RootPrivKey)
		data := tx.Encode()
		var decodedTx *AccTx
		decodedTx = decodedTx.Decode(data)
		if !reflect.DeepEqual(tx, decodedTx) {
			t.Errorf("AccTx Serialization failed (%v) vs. (%v)\n", tx, decodedTx)
		}
	}
}
