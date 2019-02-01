package protocol

import (
	"fmt"
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

func TestValidatorShardAssignment(t *testing.T) {

	//Int values of the validators
	minerA := 7076709494514296622
	minerB := -5792227536373707137
	minerC := -2069733853603621859
	minerD := -5950487042809319484
	minerE := 4049418945178417089
	minerF := -8588276547143327147
	minerG := -6646190410742499119
	minerH := -7289242113092823196
	minerI := 4178447941013997305
	minerJ := -3885344446679922789

	fmt.Printf("=== 1 Shard ===\n")
	fmt.Printf("MinerA: %d\n", int((Abs(int32(minerA)) % int32(1)) + 1))
	fmt.Printf("MinerB: %d\n", int((Abs(int32(minerB)) % int32(1)) + 1))
	fmt.Printf("MinerC: %d\n", int((Abs(int32(minerC)) % int32(1)) + 1))
	fmt.Printf("MinerD: %d\n", int((Abs(int32(minerD)) % int32(1)) + 1))
	fmt.Printf("MinerE: %d\n", int((Abs(int32(minerE)) % int32(1)) + 1))
	fmt.Printf("MinerF: %d\n", int((Abs(int32(minerF)) % int32(1)) + 1))
	fmt.Printf("MinerG: %d\n", int((Abs(int32(minerG)) % int32(1)) + 1))
	fmt.Printf("MinerH: %d\n", int((Abs(int32(minerH)) % int32(1)) + 1))
	fmt.Printf("MinerI: %d\n", int((Abs(int32(minerI)) % int32(1)) + 1))
	fmt.Printf("MinerJ: %d\n", int((Abs(int32(minerJ)) % int32(1)) + 1))


	fmt.Printf("=== 2 Shards ===\n")
	fmt.Printf("MinerA: %d\n", int((Abs(int32(minerA)) % int32(2)) + 1))
	fmt.Printf("MinerB: %d\n", int((Abs(int32(minerB)) % int32(2)) + 1))
	fmt.Printf("MinerC: %d\n", int((Abs(int32(minerC)) % int32(2)) + 1))
	fmt.Printf("MinerD: %d\n", int((Abs(int32(minerD)) % int32(2)) + 1))
	fmt.Printf("MinerE: %d\n", int((Abs(int32(minerE)) % int32(2)) + 1))
	fmt.Printf("MinerF: %d\n", int((Abs(int32(minerF)) % int32(2)) + 1))
	fmt.Printf("MinerG: %d\n", int((Abs(int32(minerG)) % int32(2)) + 1))
	fmt.Printf("MinerH: %d\n", int((Abs(int32(minerH)) % int32(2)) + 1))
	fmt.Printf("MinerI: %d\n", int((Abs(int32(minerI)) % int32(2)) + 1))
	fmt.Printf("MinerJ: %d\n", int((Abs(int32(minerJ)) % int32(2)) + 1))

	//Get Shard Assignment with 1 shard
	fmt.Printf("=== 3 Shards ===\n")
	fmt.Printf("MinerA: %d\n", int((Abs(int32(minerA)) % int32(3)) + 1))
	fmt.Printf("MinerB: %d\n", int((Abs(int32(minerB)) % int32(3)) + 1))
	fmt.Printf("MinerC: %d\n", int((Abs(int32(minerC)) % int32(3)) + 1))
	fmt.Printf("MinerD: %d\n", int((Abs(int32(minerD)) % int32(3)) + 1))
	fmt.Printf("MinerE: %d\n", int((Abs(int32(minerE)) % int32(3)) + 1))
	fmt.Printf("MinerF: %d\n", int((Abs(int32(minerF)) % int32(3)) + 1))
	fmt.Printf("MinerG: %d\n", int((Abs(int32(minerG)) % int32(3)) + 1))
	fmt.Printf("MinerH: %d\n", int((Abs(int32(minerH)) % int32(3)) + 1))
	fmt.Printf("MinerI: %d\n", int((Abs(int32(minerI)) % int32(3)) + 1))
	fmt.Printf("MinerJ: %d\n", int((Abs(int32(minerJ)) % int32(3)) + 1))

	//Get Shard Assignment with 1 shard
	fmt.Printf("=== 4 Shards ===\n")
	fmt.Printf("MinerA: %d\n", int((Abs(int32(minerA)) % int32(4)) + 1))
	fmt.Printf("MinerB: %d\n", int((Abs(int32(minerB)) % int32(4)) + 1))
	fmt.Printf("MinerC: %d\n", int((Abs(int32(minerC)) % int32(4)) + 1))
	fmt.Printf("MinerD: %d\n", int((Abs(int32(minerD)) % int32(4)) + 1))
	fmt.Printf("MinerE: %d\n", int((Abs(int32(minerE)) % int32(4)) + 1))
	fmt.Printf("MinerF: %d\n", int((Abs(int32(minerF)) % int32(4)) + 1))
	fmt.Printf("MinerG: %d\n", int((Abs(int32(minerG)) % int32(4)) + 1))
	fmt.Printf("MinerH: %d\n", int((Abs(int32(minerH)) % int32(4)) + 1))
	fmt.Printf("MinerI: %d\n", int((Abs(int32(minerI)) % int32(4)) + 1))
	fmt.Printf("MinerJ: %d\n", int((Abs(int32(minerJ)) % int32(4)) + 1))

	//Get Shard Assignment with 1 shard
	fmt.Printf("=== 5 Shards ===\n")
	fmt.Printf("MinerA: %d\n", int((Abs(int32(minerA)) % int32(5)) + 1))
	fmt.Printf("MinerB: %d\n", int((Abs(int32(minerB)) % int32(5)) + 1))
	fmt.Printf("MinerC: %d\n", int((Abs(int32(minerC)) % int32(5)) + 1))
	fmt.Printf("MinerD: %d\n", int((Abs(int32(minerD)) % int32(5)) + 1))
	fmt.Printf("MinerE: %d\n", int((Abs(int32(minerE)) % int32(5)) + 1))
	fmt.Printf("MinerF: %d\n", int((Abs(int32(minerF)) % int32(5)) + 1))
	fmt.Printf("MinerG: %d\n", int((Abs(int32(minerG)) % int32(5)) + 1))
	fmt.Printf("MinerH: %d\n", int((Abs(int32(minerH)) % int32(5)) + 1))
	fmt.Printf("MinerI: %d\n", int((Abs(int32(minerI)) % int32(5)) + 1))
	fmt.Printf("MinerJ: %d\n", int((Abs(int32(minerJ)) % int32(5)) + 1))

	//Get Shard Assignment with 1 shard
	fmt.Printf("=== 6 Shards ===\n")
	fmt.Printf("MinerA: %d\n", int((Abs(int32(minerA)) % int32(6)) + 1))
	fmt.Printf("MinerB: %d\n", int((Abs(int32(minerB)) % int32(6)) + 1))
	fmt.Printf("MinerC: %d\n", int((Abs(int32(minerC)) % int32(6)) + 1))
	fmt.Printf("MinerD: %d\n", int((Abs(int32(minerD)) % int32(6)) + 1))
	fmt.Printf("MinerE: %d\n", int((Abs(int32(minerE)) % int32(6)) + 1))
	fmt.Printf("MinerF: %d\n", int((Abs(int32(minerF)) % int32(6)) + 1))
	fmt.Printf("MinerG: %d\n", int((Abs(int32(minerG)) % int32(6)) + 1))
	fmt.Printf("MinerH: %d\n", int((Abs(int32(minerH)) % int32(6)) + 1))
	fmt.Printf("MinerI: %d\n", int((Abs(int32(minerI)) % int32(6)) + 1))
	fmt.Printf("MinerJ: %d\n", int((Abs(int32(minerJ)) % int32(6)) + 1))

	//Get Shard Assignment with 1 shard
	fmt.Printf("=== 7 Shards ===\n")
	fmt.Printf("MinerA: %d\n", int((Abs(int32(minerA)) % int32(7)) + 1))
	fmt.Printf("MinerB: %d\n", int((Abs(int32(minerB)) % int32(7)) + 1))
	fmt.Printf("MinerC: %d\n", int((Abs(int32(minerC)) % int32(7)) + 1))
	fmt.Printf("MinerD: %d\n", int((Abs(int32(minerD)) % int32(7)) + 1))
	fmt.Printf("MinerE: %d\n", int((Abs(int32(minerE)) % int32(7)) + 1))
	fmt.Printf("MinerF: %d\n", int((Abs(int32(minerF)) % int32(7)) + 1))
	fmt.Printf("MinerG: %d\n", int((Abs(int32(minerG)) % int32(7)) + 1))
	fmt.Printf("MinerH: %d\n", int((Abs(int32(minerH)) % int32(7)) + 1))
	fmt.Printf("MinerI: %d\n", int((Abs(int32(minerI)) % int32(7)) + 1))
	fmt.Printf("MinerJ: %d\n", int((Abs(int32(minerJ)) % int32(7)) + 1))
}

func Abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}