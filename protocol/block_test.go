package protocol

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestBlockCreation(t *testing.T) {
	var prevHash [32]byte
	var height uint32

	rand.Read(prevHash[:])
	height = 100

	createdBlock := NewBlock(prevHash, height)

	if !reflect.DeepEqual(createdBlock.PrevHash, prevHash) {
		t.Errorf("Previous hash does not match the given one: %x vs. %x", createdBlock.PrevHash, prevHash);
	}

	if !reflect.DeepEqual(createdBlock.Height, height) {
		t.Errorf("Height does not match the given one: %x vs. %x", createdBlock.Height, height);
	}
}

func TestBlockHash(t *testing.T) {
	var prevHash [32]byte
	var height uint32

	rand.Read(prevHash[:])
	height = 100

	block := NewBlock(prevHash, height)

	hash1 := block.HashBlock()

	if !reflect.DeepEqual(hash1, block.HashBlock()) {
		t.Errorf("Block hashing failed!")
	}

	rand.Read(prevHash[:])
	height = 101

	block.PrevHash = prevHash
	block.Height = height

	hash2 := block.HashBlock()

	if !reflect.DeepEqual(hash2, block.HashBlock()) {
		t.Errorf("Block hashing failed!")
	}
}

func TestBlockSerialization(t *testing.T) {
	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	var block Block

	block.Header = 1
	rand.Read(block.Hash[:])
	rand.Read(block.PrevHash[:])
	rand.Read(block.Nonce[:])
	block.Timestamp = time.Now().Unix()
	rand.Read(block.MerkleRoot[:])
	rand.Read(block.Beneficiary[:])
	block.NrAccTx = uint16(randVar.Uint32())
	block.NrFundsTx = uint16(randVar.Uint32())
	block.NrConfigTx = uint8(randVar.Uint32())
	block.NrStakeTx = uint16(randVar.Uint32())
	rand.Read(block.SlashedAddress[:])
	block.Height = uint32(randVar.Uint32())
	rand.Read(block.ConflictingBlockHash1[:])
	rand.Read(block.ConflictingBlockHash2[:])

	var compareBlock Block
	encodedBlock := block.Encode()
	compareBlock = *compareBlock.Decode(encodedBlock)

	if !reflect.DeepEqual(block, compareBlock) {
		t.Error("Block encoding/decoding failed!")
	}
}

func TestBlockHeaderSerialization(t *testing.T) {
	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	var blockHeader Block

	blockHeader.Header = 1
	rand.Read(blockHeader.Hash[:])
	rand.Read(blockHeader.PrevHash[:])
	blockHeader.NrConfigTx = uint8(randVar.Uint32())
	blockHeader.NrElementsBF = uint16(randVar.Uint32())

	var v1, v2, v3 [32]byte
	rand.Read(v1[:])
	rand.Read(v2[:])
	rand.Read(v3[:])

	blockHeader.InitBloomFilter([][32]byte{v1, v2, v3})

	blockHeader.Height = uint32(randVar.Uint32())
	rand.Read(blockHeader.Beneficiary[:])

	var compareBlockHeader Block
	encodedBlock := blockHeader.EncodeHeader()
	compareBlockHeader = *compareBlockHeader.Decode(encodedBlock)

	if !reflect.DeepEqual(blockHeader, compareBlockHeader) {
		t.Error("Block encoding/decoding failed!")
	}

	if blockHeader.BloomFilter.Test(v1[:]) == false {
		t.Error("Bloomfilter test failed!")
	}

	if blockHeader.BloomFilter.Test(v2[:]) == false {
		t.Error("Bloomfilter test failed!")
	}

	if blockHeader.BloomFilter.Test(v3[:]) == false {
		t.Error("Bloomfilter test failed!")
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
