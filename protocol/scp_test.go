package protocol

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestSCPSerialization(t *testing.T) {
	var merkleRootsBefore [][32]byte
	nofProofs, scp := getDummySCP()

	for i := 0; i < nofProofs; i++ {
		merkleRoot, err := scp.CalculateMerkleRootFor(i)
		if err != nil {
			t.Error("failed to calculate merkle root")
		}
		merkleRootsBefore = append(merkleRootsBefore, merkleRoot)
	}

	encoded := scp.Encode()

	var decoded *SCP
	decoded = decoded.Decode(encoded)

	for i := 0; i < nofProofs; i++ {
		if !reflect.DeepEqual(scp.Proofs[i], decoded.Proofs[i]) {
			t.Errorf("Proof does not match the given one: %v vs. %v", scp.Proofs[i].String(), decoded.Proofs[i].String())
		}

		merkleRootBefore := merkleRootsBefore[i]
		merkleRootAfter, err := decoded.CalculateMerkleRootFor(i)
		if err != nil {
			t.Error("failed to calculate merkle root")
		}

		if merkleRootBefore != merkleRootAfter {
			t.Errorf("SCP serialization failed: Merkle roots not the same \n(%v)\nvs.\n(%v)\n", merkleRootAfter, merkleRootsBefore[i])
		}
	}
}

func TestSCPHash(t *testing.T) {
	_, scp1 := getDummySCP()
	scp1Hash1 := scp1.Hash()

	if !reflect.DeepEqual(scp1Hash1, scp1.Hash()) {
		t.Errorf("SCP hashing failed!")
	}

	scp1.Proofs[0].Height++
	scp1Hash2 := scp1.Hash()

	if reflect.DeepEqual(scp1Hash1, scp1Hash2) {
		t.Errorf("SCP hashing failed!")
	}

	_, scp2 := getDummySCP()
	if reflect.DeepEqual(scp1Hash1, scp2.Hash()) {
		t.Errorf("SCP hashing failed!")
	}
}

func getDummySCP() (int, *SCP) {
	randVal := rand.New(rand.NewSource(time.Now().Unix()))
	nofProofs := int(randVal.Uint32() % 10) + 1

	scp := NewSCP()
	for i := 0; i < nofProofs; i++ {
		randHeight := (randVal.Uint32() % 10) + uint32(i * 10)
		proof := NewMerkleProof(randHeight, [][33]byte{}, 0x01, randVal.Uint64()%100000+1, randVal.Uint64()%10+1, uint32(i), accA.Address, accB.Address, nil)

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
			randVal.Read(mhash[1:33])
			proof.MHashes = append(proof.MHashes, mhash)
		}

		scp.AddMerkleProof(&proof)
	}

	return nofProofs, &scp
}