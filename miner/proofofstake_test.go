package miner

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestProofOfStake(t *testing.T) {
	cleanAndPrepare()

	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	balance := uint64(randVar.Int() % 1000)

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
	height := uint32(randVar.Int())

	diff := 10

	timestamp, _ := proofOfStake(uint8(diff), lastBlock.Hash, prevSeeds, height, balance, seed)

	if !validateProofOfStake(uint8(diff), prevSeeds, height, balance, seed, timestamp) {
		fmt.Printf("Invalid PoS calculation\n")
	}
}

func TestGetLatestSeeds(t *testing.T) {
	var testSeeds [][32]byte
	var genesisSeedSlice [32]byte
	copy(genesisSeedSlice[:], storage.GENESIS_SEED)

	b0 := newBlock([32]byte{}, genesisSeedSlice, storage.SerializeHashContent(genesisSeedSlice), 1)
	finalizeBlock(b0)
	validate(b0, false)
	testSeeds = Prepend(testSeeds, genesisSeedSlice)

	b1 := newBlock(b0.Hash, [32]byte{}, [32]byte{}, b0.Height+1)
	finalizeBlock(b1)
	validate(b1, false)
	testSeeds = Prepend(testSeeds, b1.Seed)

	b2 := newBlock(b1.Hash, [32]byte{}, [32]byte{}, b1.Height+1)
	finalizeBlock(b2)
	validate(b2, false)
	testSeeds = Prepend(testSeeds, b2.Seed)

	b3 := newBlock(b2.Hash, [32]byte{}, [32]byte{}, b2.Height+1)
	finalizeBlock(b3)
	validate(b3, false)
	testSeeds = Prepend(testSeeds, b3.Seed)

	prevSeeds := GetLatestSeeds(0, b3)

	if !reflect.DeepEqual(len(prevSeeds), 0) {
		t.Errorf("Could not retrieve the correct amount of previous seeds: %v vs. %v", len(prevSeeds), 0)
	}

	prevSeeds = GetLatestSeeds(1, b3)

	if !reflect.DeepEqual(len(prevSeeds), 1) {
		t.Errorf("Could not retrieve the correct amount of previous seeds: %v vs. %v", len(prevSeeds), 1)
	}
	if !reflect.DeepEqual(prevSeeds[len(prevSeeds)-1], b3.Seed) {
		t.Errorf("Could not retrieve the b3 seed: %x vs. %x", prevSeeds[len(prevSeeds)-1], b3.Seed)
	}

	prevSeeds = GetLatestSeeds(2, b3)

	if !reflect.DeepEqual(len(prevSeeds), 2) {
		t.Errorf("Could not retrieve the correct amount of previous seeds: %v vs. %v", len(prevSeeds), 2)
	}
	if !reflect.DeepEqual(prevSeeds[len(prevSeeds)-2], b2.Seed) {
		t.Errorf("Could not retrieve the b2 seed: %x vs. %x", prevSeeds[len(prevSeeds)-2], b2.Seed)
	}

	prevSeeds = GetLatestSeeds(3, b3)

	if !reflect.DeepEqual(len(prevSeeds), 3) {
		t.Errorf("Could not retrieve the correct amount of previous seeds: %v vs. %v", len(prevSeeds), 3)
	}
	if !reflect.DeepEqual(prevSeeds[len(prevSeeds)-3], b1.Seed) {
		t.Errorf("Could not retrieve the b1 seed: %x vs. %x", prevSeeds[len(prevSeeds)-3], b1.Seed)
	}

	prevSeeds = GetLatestSeeds(4, b3)

	if !reflect.DeepEqual(len(prevSeeds), 4) {
		t.Errorf("Could not retrieve the correct amount of previous seeds: %v vs. %v", len(prevSeeds), 4)
	}
	if !reflect.DeepEqual(prevSeeds[len(prevSeeds)-4], b0.Seed) {
		t.Errorf("Could not retrieve the b0 seed: %x vs. %x", prevSeeds[len(prevSeeds)-4], b0.Seed)
	}

	// Overflow
	prevSeeds = GetLatestSeeds(5, b3)

	if !reflect.DeepEqual(len(prevSeeds), 4) {
		t.Errorf("Retrieving more than %v previous seeds should not be possible", len(prevSeeds))
	}
}
