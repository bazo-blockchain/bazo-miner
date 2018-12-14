package miner

import (
	"reflect"
	"testing"

	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

//Tests whether state is the same before validation and after rollback of a block
func TestValidateBlockRollback(t *testing.T) {
	cleanAndPrepare()

	b := protocol.NewBlock([32]byte{}, 1)

	//Make state snapshot
	accsBefore := make(map[[64]byte]protocol.Account)
	accsBefore2 := make(map[[64]byte]protocol.Account)
	accsAfter := make(map[[64]byte]protocol.Account)

	for _, acc := range storage.State {
		accsBefore[acc.Address] = *acc
	}

	//Fill block with random transactions, finalize (PoW etc.) and validate (state change)
	createBlockWithTxs(b)
	if err := finalizeBlock(b); err != nil {
		t.Errorf("Could not finalize block: %v\n", err)
	}
	if err := validate(b, false); err != nil {
		t.Errorf("Could not validate block: %v\n", err)
	}

	for _, acc := range storage.State {
		accsAfter[acc.Address] = *acc
	}

	if reflect.DeepEqual(accsBefore, accsAfter) {
		t.Errorf("State wasn't changed despite validating a block!\n%v\n\n%v", accsBefore, accsAfter)
	}

	err := rollback(b)
	if err != nil {
		t.Errorf("%v\n", err)
	}

	for _, acc := range storage.State {
		accsBefore2[acc.Address] = *acc
	}
	accsBefore2 = resetStakingBlockHeight(accsBefore2)
	accsBefore = resetStakingBlockHeight(accsBefore)
	if !reflect.DeepEqual(accsBefore, accsBefore2) {
		t.Error("State wasn't rolled back")
	}
}

//Same test as TestValidateBlockRollback but with multiple blocks validations/rollbacks
func TestMultipleBlocksRollback(t *testing.T) {
	//Create 4 blocks after genesis, rollback 3
	cleanAndPrepare()

	//State snapshot
	stateb := make(map[[64]byte]protocol.Account)
	stateb2 := make(map[[64]byte]protocol.Account)
	stateb3 := make(map[[64]byte]protocol.Account)
	tmpState := make(map[[64]byte]protocol.Account)

	//system parameters
	var paramb []Parameters
	var paramb2 []Parameters
	var paramb3 []Parameters

	b := protocol.NewBlock([32]byte{}, 1)
	createBlockWithTxs(b)
	finalizeBlock(b)
	if err := validate(b, false); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	for _, acc := range storage.State {
		stateb[acc.Address] = *acc
	}

	paramb = make([]Parameters, len(parameterSlice))
	copy(paramb, parameterSlice)

	b2 := protocol.NewBlock(b.Hash, 2)
	createBlockWithTxs(b2)
	finalizeBlock(b2)
	if err := validate(b2, false); err != nil {
		t.Errorf("Block failed: %v\n", b2)
	}

	for _, acc := range storage.State {
		stateb2[acc.Address] = *acc
	}

	paramb2 = make([]Parameters, len(parameterSlice))
	copy(paramb2, parameterSlice)

	b3 := protocol.NewBlock(b2.Hash, 3)
	createBlockWithTxs(b3)
	finalizeBlock(b3)
	if err := validate(b3, false); err != nil {
		t.Errorf("Block failed: %v\n", b3)
	}

	for _, acc := range storage.State {
		stateb3[acc.Address] = *acc
	}

	paramb3 = make([]Parameters, len(parameterSlice))
	copy(paramb3, parameterSlice)

	b4 := protocol.NewBlock(b3.Hash, 4)
	createBlockWithTxs(b4)
	finalizeBlock(b4)
	if err := validate(b4, false); err != nil {
		t.Errorf("Block failed: %v\n", b4)
	}

	//STARTING ROLLBACKS---------------------------------------------
	if err := rollback(b4); err != nil {
		t.Errorf("%v\n", err)
	}
	for _, acc := range storage.State {
		tmpState[acc.Address] = *acc
	}
	tmpState = resetStakingBlockHeight(tmpState)
	stateb3 = resetStakingBlockHeight(stateb3)
	if !reflect.DeepEqual(tmpState, stateb3) || !reflect.DeepEqual(paramb3, parameterSlice) {
		t.Error("Block rollback failed.")
		return
	}
	//delete tmpState
	for k := range tmpState {
		delete(tmpState, k)
	}

	if err := rollback(b3); err != nil {
		t.Errorf("%v\n", err)
		return
	}
	for _, acc := range storage.State {
		tmpState[acc.Address] = *acc
	}
	tmpState = resetStakingBlockHeight(tmpState)
	stateb2 = resetStakingBlockHeight(stateb2)
	if !reflect.DeepEqual(tmpState, stateb2) || !reflect.DeepEqual(paramb2, parameterSlice) {
		t.Error("Block rollback failed.")
	}
	for k := range tmpState {
		delete(tmpState, k)
	}

	if err := rollback(b2); err != nil {
		t.Errorf("%v\n", err)
	}
	for _, acc := range storage.State {
		tmpState[acc.Address] = *acc
	}
	tmpState = resetStakingBlockHeight(tmpState)
	stateb = resetStakingBlockHeight(stateb)
	if !reflect.DeepEqual(tmpState, stateb) || !reflect.DeepEqual(paramb, parameterSlice) {
		t.Error("Block rollback failed.")
	}
	for k := range tmpState {
		delete(tmpState, k)
	}

	if err := rollback(b); err != nil {
		t.Errorf("%v\n", err)
	}
	for _, acc := range storage.State {
		tmpState[acc.Address] = *acc
	}

	for k := range tmpState {
		delete(tmpState, k)
	}
}

// resetStakingBlockHeight sets the StackingBlockHeight of all accounts to 0.
// This is needed so that the other fields can get tested.
// TODO Remove this function if rollback of StakingBlockHeight gets implemented.
func resetStakingBlockHeight(accounts map[[64]byte]protocol.Account) map[[64]byte]protocol.Account {
	accountsNoStakingBlockHeight := make(map[[64]byte]protocol.Account)

	for hash, acc := range accounts {
		acc.StakingBlockHeight = 0
		accountsNoStakingBlockHeight[hash] = acc
	}

	return accountsNoStakingBlockHeight
}
