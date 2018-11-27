package protocol

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestGenesisCreation(t *testing.T) {
	genesis := NewGenesis(accA.Address, accA.CommitmentKey)

	if !reflect.DeepEqual(genesis.RootAddress, accA.Address) {
		t.Errorf("RootAddress does not match the given one: %x vs. %x", genesis.RootAddress, accA.Address)
	}

	if !reflect.DeepEqual(genesis.RootCommitment, accA.CommitmentKey) {
		t.Errorf("RootCommitment does not match the given one: %x vs. %x", genesis.RootCommitment, accA.CommitmentKey)
	}
}

func TestGenesisSerialization(t *testing.T) {
	var genesis Genesis

	rand.Read(genesis.RootAddress[:])
	rand.Read(genesis.RootCommitment[:])

	var compareGenesis Genesis
	encodedGenesis := genesis.Encode()
	compareGenesis = *compareGenesis.Decode(encodedGenesis)

	if !reflect.DeepEqual(genesis, compareGenesis) {
		t.Error("Genesis encoding/decoding failed!")
	}
}

func TestInitialBlockWithGenesisHash(t *testing.T) {
	var genesis Genesis

	rand.Read(genesis.RootAddress[:])
	rand.Read(genesis.RootCommitment[:])

	genesisHash := genesis.Hash()

	createdBlock := NewBlock(genesisHash, 0)

	if !reflect.DeepEqual(createdBlock.PrevHash, genesis.Hash()) {
		t.Errorf("Previous hash does not match the genesis hash: %x vs. %x", createdBlock.PrevHash, genesisHash)
	}
}