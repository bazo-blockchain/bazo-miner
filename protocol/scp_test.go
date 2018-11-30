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
		merkleRoot, err := scp.CalculateMerkleRoot(i)
		if err != nil {
			t.Error("failed to calculate merkle root")
		}
		merkleRootsBefore = append(merkleRootsBefore, merkleRoot)
	}

	encoded := scp.Encode()

	var decoded *SCP
	decoded = decoded.Decode(encoded)

	if !reflect.DeepEqual(scp.Height, decoded.Height) {
		t.Errorf("Property mismatch 'Height' does not match the given one: %v vs. %v", scp.Height, decoded.Height)
	}

	if !reflect.DeepEqual(scp.MHashes, decoded.MHashes) {
		t.Errorf("Property mismatch 'MHashes' does not match the given one: %x vs. %x", scp.MHashes, decoded.MHashes)
	}

	if !reflect.DeepEqual(scp.PHeader, decoded.PHeader) {
		t.Errorf("Property mismatch 'PHeader' does not match the given one: %v vs. %v", scp.PHeader, decoded.PHeader)
	}

	if !reflect.DeepEqual(scp.PAmount, decoded.PAmount) {
		t.Errorf("Property mismatch 'PAmount' does not match the given one: %v vs. %v", scp.PAmount, decoded.PAmount)
	}

	if !reflect.DeepEqual(scp.PFee, decoded.PFee) {
		t.Errorf("Property mismatch 'PFee' does not match the given one: %v vs. %v", scp.PFee, decoded.PFee)
	}

	if !reflect.DeepEqual(scp.PTxCnt, decoded.PTxCnt) {
		t.Errorf("Property mismatch 'PTxCnt' does not match the given one: %v vs. %v", scp.PTxCnt, decoded.PTxCnt)
	}

	if !reflect.DeepEqual(scp.PFrom, decoded.PFrom) {
		t.Errorf("Property mismatch 'PFrom' does not match the given one: %v vs. %v", scp.PFrom, decoded.PFrom)
	}

	if !reflect.DeepEqual(scp.PTo, decoded.PTo) {
		t.Errorf("Property mismatch 'PTo' does not match the given one: %v vs. %v", scp.PTo, decoded.PTo)
	}

	if !reflect.DeepEqual(scp.PData, decoded.PData) {
		t.Errorf("Property mismatch 'PData' does not match the given one: %v vs. %v", scp.PData, decoded.PData)
	}

	for i := 0; i < nofProofs; i++ {
		if scp.StringFor(i) != decoded.StringFor(i) {
			t.Errorf("SCP serialization failed: Proofs not the same \n(%v)\nvs.\n(%v)\n", scp.StringFor(i), decoded.StringFor(i))
		}

		merkleRootAfter, err := decoded.CalculateMerkleRoot(i)
		if err != nil {
			t.Error("failed to calculate merkle root")
		}

		if merkleRootAfter != merkleRootsBefore[i] {
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

	scp1.Height[0] = scp1.Height[0] + 1
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

	var scp SCP
	for i := 0; i < nofProofs; i++ {
		tx, _ := ConstrFundsTx(0x01, randVal.Uint64()%100000+1, randVal.Uint64()%10+1, uint32(i), accA.Address, accB.Address, PrivKeyA, nil)

		scp.Height = append(scp.Height, (randVal.Uint32() % 10) + uint32(i * 10))
		scp.PHeader = append(scp.PHeader, tx.Header)
		scp.PAmount = append(scp.PAmount, tx.Amount)
		scp.PFee = append(scp.PFee, tx.Fee)
		scp.PTxCnt = append(scp.PTxCnt, tx.TxCnt)
		scp.PFrom = append(scp.PFrom, tx.From)
		scp.PTo = append(scp.PTo, tx.To)
		scp.PData = append(scp.PData, tx.Data)

		scp.MHashes = append(scp.MHashes, [][32]byte{})

		merkleTreeDepth := int(rand.Uint32() % 10) + 1
		for j:= 0; j < merkleTreeDepth; j++ {
			var mhash [32]byte
			randVal.Read(mhash[:])
			scp.MHashes[i] = append(scp.MHashes[i], mhash)
		}
	}

	return nofProofs, &scp
}