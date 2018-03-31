package miner

import (
	"fmt"
	"github.com/sfontanach/bazo-miner/protocol"
	"github.com/sfontanach/bazo-miner/storage"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestProofOfStake(t *testing.T) {

	cleanAndPrepare()
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	balance := uint64(rand.Int() % 1000)

	var prevSeeds [][32]byte
	prevSeed1 := protocol.CreateRandomSeed()
	prevSeeds = append(prevSeeds, prevSeed1)
	prevSeed2 := protocol.CreateRandomSeed()
	prevSeeds = append(prevSeeds, prevSeed2)
	prevSeed3 := protocol.CreateRandomSeed()
	prevSeeds = append(prevSeeds, prevSeed3)
	prevSeed4 := protocol.CreateRandomSeed()
	prevSeeds = append(prevSeeds, prevSeed4)

	seed := protocol.CreateRandomSeed()
	height := uint32(rand.Int())

	diff := 10

	timestamp, _ := proofOfStake(uint8(diff), lastBlock.Hash, prevSeeds, height, balance, seed)

	if !validateProofOfStake(uint8(diff), prevSeeds, height, balance, seed, timestamp) {
		fmt.Printf("Invalid PoS calculation\n")
	}
}

func TestGetLatestSeeds(t *testing.T) {
	cleanAndPrepare()
	var seeds [][32]byte
	var genesisSeedSlice [32]byte
	copy(genesisSeedSlice[:], storage.GENESIS_SEED)
	seeds = Prepend(seeds, genesisSeedSlice)
	//Initially we expect only the genesis seed

	b := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)

	prevSeeds := GetLatestSeeds(1, b)

	if !reflect.DeepEqual(prevSeeds[0], genesisSeedSlice) {
		t.Error("Could not retrieve the genesis seed.", prevSeeds[0], genesisSeedSlice)
	}
	if !reflect.DeepEqual(1, len(prevSeeds)) {
		t.Error("Could not retrieve the correct amount of previous seeds (all seeds).", 1, len(prevSeeds))
	}

	//Two new blocks are added with random seeds
	b1 := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	if err := finalizeBlock(b1); err != nil {
		t.Error("Error finalizing b1", err)
	}
	seeds = Prepend(seeds, b1.Seed)
	validateBlock(b1)

	b2 := newBlock(b1.Hash, [32]byte{}, [32]byte{}, b1.Height+1)
	if err := finalizeBlock(b2); err != nil {
		t.Error("Error finalizing b2", err)
	}
	validateBlock(b2)
	seeds = Prepend(seeds, b2.Seed)

	b3 := newBlock(b2.Hash, [32]byte{}, [32]byte{}, b2.Height+1)

	prevSeeds = GetLatestSeeds(3, b3)

	//Two new blocks are added with random seeds
	if !reflect.DeepEqual(prevSeeds, seeds) {
		t.Errorf("Could not retrieve previous seeds correctly (all seeds).\n%v\n%v", prevSeeds, seeds)
	}
	if !reflect.DeepEqual(3, len(prevSeeds)) {
		t.Error("Could not retrieve the correct amount of previous seeds (all seeds).", 3, len(prevSeeds))
	}

	prevSeeds = GetLatestSeeds(2, b3)

	if !reflect.DeepEqual(prevSeeds, seeds[0:2]) {
		t.Error("Could not retrieve previous seeds correctly (n < block height).", prevSeeds, seeds[0:2])
	}
	if !reflect.DeepEqual(2, len(prevSeeds)) {
		t.Error("Could not retrieve the correct amount of previous seeds  (n < block height).", 2, len(prevSeeds))
	}

	//5 seeds are expected since only 5 blocks are in the blockchain
	prevSeeds = GetLatestSeeds(5, b3)

	if !reflect.DeepEqual(prevSeeds, seeds[0:3]) {
		t.Errorf("Could not retrieve previous seeds correctly (all seeds).\n%x\n%x", prevSeeds, seeds[0:3])
	}
	if !reflect.DeepEqual(3, len(prevSeeds)) {
		t.Error("Could not retrieve the correct amount of previous seeds (n > block height).", 3, len(prevSeeds))
	}
}
