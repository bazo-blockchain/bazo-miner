package protocol

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestMerkleProofSerialization(t *testing.T) {
	randVal := rand.New(rand.NewSource(time.Now().Unix()))
	randHeight := randVal.Uint32() % 10
	var address AddressType
	var merkleRoot HashType
	randVal.Read(address[:])
	randVal.Read(merkleRoot[:])

	proof := NewMerkleProof(randHeight, [][33]byte{}, address, randVal.Int63()%100000+1, merkleRoot)
	merkleTreeDepth := int(rand.Uint32() % 10) + 1
	for j:= 0; j < merkleTreeDepth; j++ {
		leftOrRightNumber := int(rand.Uint32() % 2)

		var mhash [33]byte
		var leftOrRight [1]byte
		if leftOrRightNumber == 0 {
			leftOrRight = [1]byte{'l'}
		} else {
			leftOrRight = [1]byte{'r'}
		}

		copy(mhash[0:1], leftOrRight[:])
		rand.Read(mhash[1:33])
		proof.MerkleHashes = append(proof.MerkleHashes, mhash)
	}

	merkleRootBefore, err := proof.CalculateMerkleRoot()
	if err != nil {
		t.Error(err)
	}

	encoded := proof.Encode()

	var decoded *MerkleProof
	decoded = decoded.Decode(encoded)

	if !reflect.DeepEqual(&proof, decoded) {
		t.Errorf("Proof does not match the given one: \n%vvs.\n%v", proof.String(), decoded.String())
	}

	merkleRootAfter, err := decoded.CalculateMerkleRoot()
	if err != nil {
		t.Error("failed to calculate merkle root")
	}

	if merkleRootBefore != merkleRootAfter {
		t.Errorf("SCP serialization failed: Merkle roots not the same \n(%v)\nvs.\n(%v)\n", merkleRootBefore, merkleRootAfter)
	}
}