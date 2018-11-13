package miner

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
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
	proofs = append([][crypto.COMM_PROOF_LENGTH]byte{genesisCommitmentProof}, proofs...)
	//Initially we expect only the genesis commitment proof

	b := newBlock([32]byte{}, [crypto.COMM_PROOF_LENGTH]byte{}, 1)

	prevProofs := GetLatestProofs(1, b)

	if !reflect.DeepEqual(prevProofs[0], genesisCommitmentProof) {
		t.Error("Could not retrieve the genesis commitment proof.", prevProofs[0], genesisCommitmentProof)
	}
	if !reflect.DeepEqual(1, len(prevProofs)) {
		t.Error("Could not retrieve the correct amount of previous proofs (all proofs).", 1, len(prevProofs))
	}

	//Two new blocks are added with random commitment proofs
	b1 := newBlock([32]byte{}, [crypto.COMM_PROOF_LENGTH]byte{}, 1)
	if err := finalizeBlock(b1); err != nil {
		t.Error("Error finalizing b1", err)
	}
	proofs = append([][crypto.COMM_PROOF_LENGTH]byte{b1.CommitmentProof}, proofs...)
	validate(b1, false)

	b2 := newBlock(b1.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, b1.Height+1)
	if err := finalizeBlock(b2); err != nil {
		t.Error("Error finalizing b2", err)
	}
	validate(b2, false)
	proofs = append([][crypto.COMM_PROOF_LENGTH]byte{b2.CommitmentProof}, proofs...)

	b3 := newBlock(b2.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, b2.Height+1)

	prevProofs = GetLatestProofs(3, b3)

	//Two new blocks are added with random commitment proofs
	if !reflect.DeepEqual(prevProofs, proofs) {
		t.Errorf("Could not retrieve previous proofs correctly (all proofs).\n%v\n%v", prevProofs, proofs)
	}
	if !reflect.DeepEqual(3, len(prevProofs)) {
		t.Error("Could not retrieve the correct amount of previous proofs (all proofs).", 3, len(prevProofs))
	}

	prevProofs = GetLatestProofs(2, b3)

	if !reflect.DeepEqual(prevProofs, proofs[0:2]) {
		t.Error("Could not retrieve previous proofs correctly (n < block height).", prevProofs, proofs[0:2])
	}
	if !reflect.DeepEqual(2, len(prevProofs)) {
		t.Error("Could not retrieve the correct amount of previous proofs  (n < block height).", 2, len(prevProofs))
	}

	//5 proofs are expected since only 5 blocks are in the blockchain
	prevProofs = GetLatestProofs(5, b3)

	if !reflect.DeepEqual(prevProofs, proofs[0:3]) {
		t.Errorf("Could not retrieve previous proofs correctly (all proofs).\n%x\n%x", prevProofs, proofs[0:3])
	}
	if !reflect.DeepEqual(3, len(prevProofs)) {
		t.Error("Could not retrieve the correct amount of previous proofs (n > block height).", 3, len(prevProofs))
	}
}
