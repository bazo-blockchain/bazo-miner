package protocol

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestConfigTxSerialization(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	loopMax := int(rand.Uint32() % 10000)
	for i := 0; i < loopMax; i++ {
		tx, err := ConstrConfigTx(uint8(rand.Uint32()%256), uint8(rand.Uint32()%256), rand.Uint64(), rand.Uint64(), uint8(i), RootPrivKey)
		data := tx.Encode()
		var decodedTx *ConfigTx
		decodedTx = decodedTx.Decode(data)
		if !reflect.DeepEqual(tx, decodedTx) || err != nil {
			t.Errorf("ConfigTx Serialization failed (%v) vs. (%v)\n", tx, decodedTx)
		}
	}
}
