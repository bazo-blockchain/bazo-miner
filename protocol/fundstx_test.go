package protocol

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestFundsTxSerialization(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	loopMax := int(rand.Uint32() % 10000)
	for i := 0; i < loopMax; i++ {
		tx, _ := NewSignedFundsTx(0x01, rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accA.Address, accB.Address, PrivKeyA, nil)

		var merkleRootsBefore [][32]byte
		proofs := getDummyProofs()

		for i := 0; i < len(proofs); i++ {
			merkleRoot, err := proofs[i].CalculateMerkleRoot()
			if err != nil {
				t.Error("failed to calculate merkle root")
			}
			merkleRootsBefore = append(merkleRootsBefore, merkleRoot)
		}

		tx.Proofs = proofs

		data := tx.Encode()
		var decodedTx *FundsTx
		decodedTx = decodedTx.Decode(data)

		//this is done by verify() which is outside protocol package, we're just testing serialization here
		decodedTx.From = accA.Address
		decodedTx.To = accB.Address

		if !reflect.DeepEqual(tx, decodedTx) {
			t.Errorf("FundsTx Serialization failed (%v) vs. (%v)\n", tx, decodedTx)
		}

		for i := 0; i < len(proofs); i++ {
			if !reflect.DeepEqual(decodedTx.Proofs[i], tx.Proofs[i]) {
				t.Errorf("Proof does not match the given one: %v vs. %v", decodedTx.Proofs[i].String(), tx.Proofs[i].String())
			}

			merkleRootBefore := merkleRootsBefore[i]
			merkleRootAfter, err := decodedTx.Proofs[i].CalculateMerkleRoot()
			if err != nil {
				t.Error("failed to calculate merkle root")
			}

			if merkleRootBefore != merkleRootAfter {
				t.Errorf("Merkle proof serialization failed: Merkle roots not the same \n(%v)\nvs.\n(%v)\n", merkleRootAfter, merkleRootsBefore[i])
			}
		}
	}
}

func getDummyProofs() (proofs []*MerkleProof) {
	randVal := rand.New(rand.NewSource(time.Now().Unix()))
	nofProofs := int(randVal.Uint32() % 10) + 1

	for i := 0; i < nofProofs; i++ {
		randHeight := (randVal.Uint32() % 10) + uint32(i * 10)
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
			randVal.Read(mhash[1:33])
			proof.MerkleHashes = append(proof.MerkleHashes, mhash)
		}

		proofs = append(proofs, &proof)
	}

	return proofs
}