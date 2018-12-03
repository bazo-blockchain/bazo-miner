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
	proof := NewMerkleProof(randHeight, [][32]byte{}, 0x01, randVal.Uint64()%100000+1, randVal.Uint64()%10+1, uint32(0), accA.Address, accB.Address, nil)

	merkleTreeDepth := int(rand.Uint32() % 10) + 1
	for j:= 0; j < merkleTreeDepth; j++ {
		var mhash [32]byte
		randVal.Read(mhash[:])
		proof.MHashes = append(proof.MHashes, mhash)
	}

	encoded := proof.Encode()

	var decoded *MerkleProof
	decoded = decoded.Decode(encoded)

	if !reflect.DeepEqual(proof, decoded) {
		t.Errorf("Proof does not match the given one: %v vs. %v", proof.String(), decoded.String())
	}

	merkleRootBefore, err := proof.CalculateMerkleRoot()
	if err != nil {
		t.Error(err)
	}

	merkleRootAfter, err := decoded.CalculateMerkleRoot()
	if err != nil {
		t.Error("failed to calculate merkle root")
	}

	if merkleRootBefore != merkleRootAfter {
		t.Errorf("SCP serialization failed: Merkle roots not the same \n(%v)\nvs.\n(%v)\n", merkleRootBefore, merkleRootAfter)
	}
}