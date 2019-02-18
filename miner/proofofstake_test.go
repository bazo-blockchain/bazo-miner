package miner

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
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

	var prevProofs [][crypto.COMM_PROOF_LENGTH]byte
	prevProof1, _ := crypto.SignMessageWithRSAKey(CommPrivKeyAccA, "0")
	prevProofs = append(prevProofs, prevProof1)
	prevProof2, _ := crypto.SignMessageWithRSAKey(CommPrivKeyAccA, "1")
	prevProofs = append(prevProofs, prevProof2)
	prevProof3, _ := crypto.SignMessageWithRSAKey(CommPrivKeyAccA, "2")
	prevProofs = append(prevProofs, prevProof3)
	prevProof4, _ := crypto.SignMessageWithRSAKey(CommPrivKeyAccA, "3")
	prevProofs = append(prevProofs, prevProof4)

	var height uint32 = 4
	diff := 10

	commitmentProof, _ := crypto.SignMessageWithRSAKey(CommPrivKeyAccA, fmt.Sprint(height))
	timestamp, _ := proofOfStake(uint8(diff), lastBlock.Hash, prevProofs, height, balance, commitmentProof)

	if !validateProofOfStake(uint8(diff), prevProofs, height, balance, commitmentProof, timestamp) {
		fmt.Printf("Invalid PoS calculation\n")
	}
}

func TestGetLatestProofs(t *testing.T) {
	cleanAndPrepare()

	var proofs [][crypto.COMM_PROOF_LENGTH]byte
	genesisCommitmentProof, _ := crypto.SignMessageWithRSAKey(CommPrivKeyRoot, "0")
	//proofs = append([][crypto.COMM_PROOF_LENGTH]byte{genesisCommitmentProof}, proofs...)
	//Initially we expect only the genesis commitment proof

	//b := newBlock([32]byte{}, [crypto.COMM_PROOF_LENGTH]byte{}, 1)
	b := newBlock(lastBlock.HashBlock(), [crypto.COMM_PROOF_LENGTH]byte{}, 2)
	if err := finalizeBlock(b); err != nil {
		t.Error("Error finalizing b1", err)
	}
	storage.WriteClosedBlock(b)
	proofs = append([][crypto.COMM_PROOF_LENGTH]byte{b.CommitmentProof}, proofs...)

	prevProofs := GetLatestProofs(1, b)

	if !reflect.DeepEqual(prevProofs[0], initialBlock.CommitmentProof) {
		t.Error("Could not retrieve the initial block commitment proof.", prevProofs[0], genesisCommitmentProof)
	}
	if !reflect.DeepEqual(1, len(prevProofs)) {
		t.Error("Could not retrieve the correct amount of previous proofs (all proofs).", 1, len(prevProofs))
	}

	//Two new blocks are added with random commitment proofs
	//b1 := newBlock([32]byte{}, [crypto.COMM_PROOF_LENGTH]byte{}, 1)
	lastBlock = b
	b1 := newBlock(b.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, 3)
	if err := finalizeBlock(b1); err != nil {
		t.Error("Error finalizing b1", err)
	}
	proofs = append([][crypto.COMM_PROOF_LENGTH]byte{b1.CommitmentProof}, proofs...)
	validate(b1, false)
	storage.WriteClosedBlock(b1)
	lastBlock = b1

	b2 := newBlock(b1.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, b1.Height+1)
	if err := finalizeBlock(b2); err != nil {
		t.Error("Error finalizing b2", err)
	}
	validate(b2, false)
	proofs = append([][crypto.COMM_PROOF_LENGTH]byte{b2.CommitmentProof}, proofs...)

	t.Log("Proofs Slice:\n")
	t.Logf("%v",CommitmentProofSliceToString(proofs))

	storage.WriteClosedBlock(b2)

	b3 := newBlock(b2.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, b2.Height+1)
	storage.WriteClosedBlock(b3)

	prevProofs = GetLatestProofs(3, b3)
	t.Log("prevProofs Slice with n=3 :\n")
	t.Logf("%v",CommitmentProofSliceToString(prevProofs))

	//Two new blocks are added with random commitment proofs
	if !reflect.DeepEqual(prevProofs, proofs) {
		t.Errorf("Could not retrieve previous proofs correctly (all proofs).\n%v\n%v", prevProofs, proofs)
	}
	if !reflect.DeepEqual(3, len(prevProofs)) {
		t.Error("Could not retrieve the correct amount of previous proofs (all proofs).", 3, len(prevProofs))
	}

	prevProofs = GetLatestProofs(2, b3)

	if !reflect.DeepEqual(prevProofs, proofs[:2]) {
		t.Errorf("Could not retrieve previous proofs correctly (n < block height)\n prevProofs: %v\nProofs: %v\n",
			CommitmentProofSliceToString(prevProofs), CommitmentProofSliceToString(proofs[1:]))
	}
	if !reflect.DeepEqual(2, len(prevProofs)) {
		t.Error("Could not retrieve the correct amount of previous proofs  (n < block height).", 2, len(prevProofs))
	}
}
