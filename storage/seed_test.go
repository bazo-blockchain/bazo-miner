package storage

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

//Write and Read test on the seed json store
func TestReadWriteJson(t *testing.T) {
	// create Seed and Hash
	seed := createRandomSeed()
	hashedSeed := protocol.SerializeHashContent(seed)

	err := AppendNewSeed(TestSeedFileName, SeedJson{fmt.Sprintf("%x", string(hashedSeed[:])), string(seed[:])})

	if err != nil {
		t.Errorf("Writing to JSON file failed.")
	}

	seedFromFile, err := GetSeed(hashedSeed, TestSeedFileName)

	if err != nil {
		t.Errorf("JSON Serialization failed.")
	}

	//compare the seed from the file with the randomly generated one
	if !reflect.DeepEqual(seed, seedFromFile) {
		t.Errorf("JSON Serialization failed (%v) vs. (%v)\n", seed, seedFromFile)
	}
}

func createRandomSeed() [32]byte {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seed [32]byte
	for i := range seed {
		seed[i] = chars[r.Intn(len(chars))]
	}
	return seed
}
