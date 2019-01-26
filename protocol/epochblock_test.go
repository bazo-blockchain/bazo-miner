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
	numberofshards := 2

	valMapping := NewMapping()
	valMapping.ValMapping[[64]byte{'0','1','2','3'}] = 1
	valMapping.ValMapping[[64]byte{'4','5','6','7'}] = 2

	createdEpochBlock := NewEpochBlock(prevShardHashes, height)
	createdEpochBlock.ValMapping = valMapping
	createdEpochBlock.NofShards = numberofshards

	if !reflect.DeepEqual(createdEpochBlock.PrevShardHashes, prevShardHashes) {
		t.Errorf("Previous hash does not match the given one: %x vs. %x", createdEpochBlock.PrevShardHashes, prevShardHashes)
	}

	if !reflect.DeepEqual(createdEpochBlock.Height, height) {
		t.Errorf("Height does not match the given one: %x vs. %x", createdEpochBlock.Height, height)
	}

	if !reflect.DeepEqual(createdEpochBlock.ValMapping, valMapping) {
		t.Errorf("Validator Shard Mapping does not match the given one: %s vs. %s", createdEpochBlock.ValMapping.String(), valMapping.String())
	}

	if !reflect.DeepEqual(createdEpochBlock.NofShards, numberofshards) {
		t.Errorf("Number of Shards does not match the given one: %s vs. %s", createdEpochBlock.ValMapping.String(), valMapping.String())
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
	numberofshards := 2

	valMapping := NewMapping()
	valMapping.ValMapping[[64]byte{'0','1','2','3'}] = 1
	valMapping.ValMapping[[64]byte{'4','5','6','7'}] = 2

	epochBlock := NewEpochBlock(prevShardHashes, height)
	epochBlock.ValMapping = valMapping
	epochBlock.NofShards = numberofshards

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
	numberofshards := 2
	stateMapping := make(map[[64]byte]*Account)

	acc1 := new(Account)
	acc1.Address = [64]byte{'1'}
	acc1.Balance = 1000

	stateMapping[[64]byte{'1'}] = acc1

	acc2 := new(Account)
	acc2.Address = [64]byte{'2'}
	acc2.Balance = 2000

	stateMapping[[64]byte{'2'}] = acc2

	acc3 := new(Account)
	acc3.Address = [64]byte{'3'}
	acc3.Balance = 3000

	stateMapping[[64]byte{'3'}] = acc3

	valMapping := NewMapping()
	valMapping.ValMapping[[64]byte{'0','1','2','3'}] = 1
	valMapping.ValMapping[[64]byte{'4','5','6','7'}] = 2

	var epochBlock EpochBlock

	epochBlock.Header = 1
	epochBlock.Hash = [32]byte{'0', '1'}
	epochBlock.PrevShardHashes = prevShardHashes
	epochBlock.Height = height
	epochBlock.MerkleRoot = [32]byte{'0', '1'}
	epochBlock.MerklePatriciaRoot = [32]byte{'0', '1'}
	epochBlock.Timestamp = time.Now().Unix()
	epochBlock.State = stateMapping
	epochBlock.ValMapping = valMapping
	epochBlock.NofShards = numberofshards

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
