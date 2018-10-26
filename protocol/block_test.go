package protocol

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestBlockSerialization(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	b := new(Block)
	b.Hash = [32]byte{0, 1, 2, 3, 4}
	b.PrevHash = [32]byte{1, 2, 3, 4, 5}
	b.Nonce = [8]byte{0, 1, 2, 3, 4, 5, 6, 7}
	b.Timestamp = time.Now().Unix()
	b.MerkleRoot = [32]byte{2, 3, 4, 5, 6}
	b.Beneficiary = [32]byte{3, 4, 5, 6, 7}
	b.NrAccTx = uint16(rand.Uint32())
	b.NrFundsTx = uint16(rand.Uint32())
	b.NrConfigTx = uint8(rand.Uint32())
	b.NrStakeTx = uint16(rand.Uint32())
	b.SlashedAddress = [32]byte{0, 1, 2, 3, 4}
	b.Height = uint32(rand.Uint32())
	b.CommitmentProof = [256]byte{0, 1, 2, 3, 4}
	b.ConflictingBlockHash1 = [32]byte{0, 1, 2, 3, 4}
	b.ConflictingBlockHash2 = [32]byte{0, 1, 2, 3, 4}

	//TODO Bloomfilter serialization

	encodedBlock := b.Encode()
	b2 := b.Decode(encodedBlock)

	if !reflect.DeepEqual(encodedBlock, b2.Encode()) {
		t.Error("Block encoding/decoding failed\n", b, b2)
	}
}

func TestGetSize(t *testing.T) {
	b := new(Block)

	b.NrAccTx = uint16(rand.Uint32())
	b.NrFundsTx = uint16(rand.Uint32())
	b.NrConfigTx = uint8(rand.Uint32())
	b.NrStakeTx = uint16(rand.Uint32())

	txAmount := b.NrAccTx + b.NrFundsTx + uint16(b.NrConfigTx) + b.NrStakeTx

	if b.GetSize() != uint64(txAmount)*HASH_LEN+128+4+MIN_BLOCK_SIZE {
		fmt.Printf("Miscalculated block size: %v vs. %v\n", b.GetSize(), uint64(txAmount)*32+MIN_BLOCK_SIZE)
	}
}
