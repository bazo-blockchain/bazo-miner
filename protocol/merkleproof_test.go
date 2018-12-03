package protocol

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestMerkleProofSerialization(t *testing.T) {
	proof := NewMerkleProof(rand.Uint32() % 10, [][33]byte{}, 0x01, rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(0), accA.Address, accB.Address, nil)
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
		proof.MHashes = append(proof.MHashes, mhash)
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