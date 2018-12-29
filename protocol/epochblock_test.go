package protocol

import (
	"reflect"
	"testing"
	"time"
)

func TestEpochBlockCreation(t *testing.T) {
	var prevShardHashes [][32]byte
	var height uint32

	//Assuming that the previous epoch had 5 running shards. Each hashX denotes the hash value of the last shard block
	hash1 := [32]byte{'0', '1'}
	hash2 := [32]byte{'0', '1'}
	hash3 := [32]byte{'0', '1'}
	hash4 := [32]byte{'0', '1'}
	hash5 := [32]byte{'0', '1'}

	prevShardHashes = append(prevShardHashes, hash1)
	prevShardHashes = append(prevShardHashes, hash2)
	prevShardHashes = append(prevShardHashes, hash3)
	prevShardHashes = append(prevShardHashes, hash4)
	prevShardHashes = append(prevShardHashes, hash5)

	height = 100

	createdEpochBlock := NewEpochBlock(prevShardHashes, height)

	if !reflect.DeepEqual(createdEpochBlock.PrevShardHashes, prevShardHashes) {
		t.Errorf("Previous hash does not match the given one: %x vs. %x", createdEpochBlock.PrevShardHashes, prevShardHashes)
	}

	if !reflect.DeepEqual(createdEpochBlock.Height, height) {
		t.Errorf("Height does not match the given one: %x vs. %x", createdEpochBlock.Height, height)
	}
}

func TestEpochBlockHash(t *testing.T) {
	var prevShardHashes [][32]byte
	var height uint32

	//Assuming that the previous epoch had 5 running shards. Each hashX denotes the hash value of the last shard block
	hash1 := [32]byte{'0', '1'}
	hash2 := [32]byte{'0', '1'}
	hash3 := [32]byte{'0', '1'}
	hash4 := [32]byte{'0', '1'}
	hash5 := [32]byte{'0', '1'}

	prevShardHashes = append(prevShardHashes, hash1)
	prevShardHashes = append(prevShardHashes, hash2)
	prevShardHashes = append(prevShardHashes, hash3)
	prevShardHashes = append(prevShardHashes, hash4)
	prevShardHashes = append(prevShardHashes, hash5)

	height = 100

	epochBlock := NewEpochBlock(prevShardHashes, height)

	hashEpoch := epochBlock.HashEpochBlock()

	if !reflect.DeepEqual(hashEpoch, epochBlock.HashEpochBlock()) {
		t.Errorf("Block hashing failed!")
	}
}

func TestEpochBlockSerialization(t *testing.T) {
	var prevShardHashes [][32]byte
	var height uint32

	//Assuming that the previous epoch had 5 running shards. Each hashX denotes the hash value of the last shard block
	hash1 := [32]byte{'0', '1'}
	hash2 := [32]byte{'0', '1'}
	hash3 := [32]byte{'0', '1'}
	hash4 := [32]byte{'0', '1'}
	hash5 := [32]byte{'0', '1'}

	prevShardHashes = append(prevShardHashes, hash1)
	prevShardHashes = append(prevShardHashes, hash2)
	prevShardHashes = append(prevShardHashes, hash3)
	prevShardHashes = append(prevShardHashes, hash4)
	prevShardHashes = append(prevShardHashes, hash5)

	height = 100

	var epochBlock EpochBlock

	epochBlock.Header = 1
	epochBlock.Hash = [32]byte{'0', '1'}
	epochBlock.PrevShardHashes = prevShardHashes
	epochBlock.Height = height
	epochBlock.MerkleRoot = [32]byte{'0', '1'}
	epochBlock.MerklePatriciaRoot = [32]byte{'0', '1'}
	epochBlock.Timestamp = time.Now().Unix()

	var compareBlock EpochBlock
	encodedBlock := epochBlock.Encode()
	compareBlock = *compareBlock.Decode(encodedBlock)

	if !reflect.DeepEqual(epochBlock, compareBlock) {
		t.Error("Block encoding/decoding failed!")
	}
}

func TestEpochBlockHeaderSerialization(t *testing.T) {
	var prevShardHashes [][32]byte
	var height uint32

	//Assuming that the previous epoch had 5 running shards. Each hashX denotes the hash value of the last shard block
	hash1 := [32]byte{'0', '1'}
	hash2 := [32]byte{'0', '1'}
	hash3 := [32]byte{'0', '1'}
	hash4 := [32]byte{'0', '1'}
	hash5 := [32]byte{'0', '1'}

	prevShardHashes = append(prevShardHashes, hash1)
	prevShardHashes = append(prevShardHashes, hash2)
	prevShardHashes = append(prevShardHashes, hash3)
	prevShardHashes = append(prevShardHashes, hash4)
	prevShardHashes = append(prevShardHashes, hash5)

	height = 100

	var epochBlockHeader EpochBlock

	epochBlockHeader.Header = 1
	epochBlockHeader.Hash = [32]byte{'0', '1'}
	epochBlockHeader.PrevShardHashes = prevShardHashes
	epochBlockHeader.Height = height


	var compareEpochBlockHeader EpochBlock
	encodedBlock := epochBlockHeader.EncodeHeader()
	compareEpochBlockHeader = *compareEpochBlockHeader.Decode(encodedBlock)

	if !reflect.DeepEqual(epochBlockHeader, compareEpochBlockHeader) {
		t.Error("Block encoding/decoding failed!")
	}
}
