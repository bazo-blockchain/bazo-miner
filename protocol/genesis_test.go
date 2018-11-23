package protocol

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestGenesisCreation(t *testing.T) {
	genesis := NewGenesis(accA.Address, accA.Address, accA.CommitmentKey)

	if !reflect.DeepEqual(genesis.RootAddress, accA.Address) {
		t.Errorf("RootAddress does not match the given one: %x vs. %x", genesis.RootAddress, accA.Address)
	}

	if !reflect.DeepEqual(genesis.RootMultisig, accA.Address) {
		t.Errorf("RootMultisig does not match the given one: %x vs. %x", genesis.RootMultisig, accA.Address)
	}

	if !reflect.DeepEqual(genesis.RootCommitment, accA.CommitmentKey) {
		t.Errorf("RootCommitment does not match the given one: %x vs. %x", genesis.RootCommitment, accA.CommitmentKey)
	}
}

func TestGenesisSerialization(t *testing.T) {
	var genesis Genesis

	rand.Read(genesis.RootAddress[:])
	rand.Read(genesis.RootCommitment[:])
	rand.Read(genesis.RootMultisig[:])

	var compareGenesis Genesis
	encodedGenesis := genesis.Encode()
	compareGenesis = *compareGenesis.Decode(encodedGenesis)

	if !reflect.DeepEqual(genesis, compareGenesis) {
		t.Error("Genesis encoding/decoding failed!")
	}
}