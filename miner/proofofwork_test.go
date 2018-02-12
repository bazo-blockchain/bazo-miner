package miner

import (
	"fmt"
	"github.com/mchetelat/bazo_miner/protocol"
	"golang.org/x/crypto/sha3"
	"math/rand"
	"testing"
	"time"
)

func TestProofOfWork(t *testing.T) {

	cleanAndPrepare()
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	//Calculate random partial hash
	partialHash := protocol.SerializeHashContent(rand.Uint32())
	diff := 10

	nonce, _ := proofOfWork(uint8(diff), partialHash, [32]byte{})

	if !validateProofOfWork(uint8(diff), sha3.Sum256(append(nonce[:], partialHash[:]...))) {
		fmt.Printf("Invalid PoW calculation\n")
	}
}

func TestValidateProofOfWork(t *testing.T) {

	var hash [32]byte

	for cnt := 0; cnt < 32; cnt++ {
		if cnt >= 8 {
			//0x3f == 0011 1111
			hash[cnt] = 0x3f
		}
	}

	//First 8*8+2 bits are set to 0
	if !validateProofOfWork(8*8+2, hash) {
		t.Error("Invalid PoW validation")
	}

	if validateProofOfWork(8*8+3, hash) {
		t.Error("Invalid PoW validation")
	}
}
