package miner

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"reflect"
	"testing"
)

func TestSlashingCondition(t *testing.T) {
	cleanAndPrepare()

	myAcc, _ := storage.GetAccount(crypto.SerializeHashContent(validatorAccAddress))
	initBalance := myAcc.Balance

	forkBlock := newBlock([32]byte{}, [crypto.COMM_PROOF_LENGTH]byte{}, 1)
	if err := finalizeBlock(forkBlock); err != nil {
		t.Errorf("Block finalization for b1 (%v) failed: %v\n", forkBlock, err)
	}
	if err := validate(forkBlock, false); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", forkBlock, err)
	}

	// genesis <- forkBlock <- b
	b := newBlock(forkBlock.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, 2)
	if err := finalizeBlock(b); err != nil {
		t.Errorf("Block finalization for b1 (%v) failed: %v\n", b, err)
	}
	if err := validate(b, false); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	//reference to an old block
	lastBlock = forkBlock

	// genesis <- forkBlock <- b2
	b2 := newBlock(forkBlock.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, 2)
	if err := finalizeBlock(b2); err != nil {
		t.Errorf("Block finalization for b2 (%v) failed: %v\n", b2, err)
	}

	//t.Logf("\ninit block:%v\nb1:%v\nb2:%v\n", forkBlock.Hash, b.Hash, b2.Hash)
	if err := validate(b2, false); err != nil {
		t.Errorf("Block validation for b2 (%v) failed: %v\n", b2, err)
	}

	slashingDict2 := make(map[[32]byte]SlashingProof)
	slashingDict2[b.Beneficiary] = SlashingProof{b2.Hash, b.Hash}

	if !reflect.DeepEqual(slashingDict, slashingDict2) {
		t.Error("Slashing dictionary was not built correctly.", slashingDict, slashingDict2)
	}

	//third block contains the slashing proof
	b3 := newBlock(b2.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, 3)
	if err := finalizeBlock(b3); err != nil {
		t.Errorf("Block finalization for b3 (%v) failed: %v\n", b3, err)
	}

	//Check whether the right proof was included in b3
	slashingDict3 := make(map[[32]byte]SlashingProof)
	slashingDict3[b3.Beneficiary] = SlashingProof{b3.ConflictingBlockHash1, b3.ConflictingBlockHash2}

	if !reflect.DeepEqual(slashingDict, slashingDict3) {
		t.Error("Slashing proof was not correctly included in b3.", slashingDict, slashingDict3)
	}

	if err := validate(b3, false); err != nil {
		t.Errorf("Block validation for b3 (%v) failed: %v\n", b3, err)
	}

	//Check whether the slashing reward is added after a slashing proof is provided
	expectedBalance := initBalance+4*activeParameters.Block_reward+activeParameters.Slash_reward-activeParameters.Staking_minimum
	if !reflect.DeepEqual(expectedBalance, myAcc.Balance) {
		t.Error("Slashing reward is not properly added.", initBalance, myAcc.Balance, expectedBalance)
	}
}
