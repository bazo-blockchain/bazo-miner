package protocol

import (
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"reflect"
	"testing"
)

func TestStateStashSetMethod(t *testing.T) {
	var sampleStash = NewStateStash()
	var relAccounts1 = make(map[[64]byte]*RelativeAccount)
	var relAccounts2 = make(map[[64]byte]*RelativeAccount)
	var relAccounts3 = make(map[[64]byte]*RelativeAccount)

	//Account information in the relative state of transition 1
	accARelState := NewRelativeAccount([64]byte{'0'},[64]byte{},-10,true,[crypto.COMM_KEY_LENGTH]byte{},nil,nil)
	accARelState.TxCnt = 9
	accARelState.StakingBlockHeight = 10
	relAccounts1[[64]byte{'0'}] = &accARelState


	accBRelState := NewRelativeAccount([64]byte{'1'},[64]byte{},10,true,[crypto.COMM_KEY_LENGTH]byte{},nil,nil)
	accBRelState.TxCnt = 9
	accBRelState.StakingBlockHeight = 10
	relAccounts1[[64]byte{'1'}] = &accBRelState

	accCRelState := NewRelativeAccount([64]byte{'2'},[64]byte{},0,true,[crypto.COMM_KEY_LENGTH]byte{},nil,nil)
	accCRelState.TxCnt = 9
	accCRelState.StakingBlockHeight = 10
	relAccounts2[[64]byte{'2'}] = &accCRelState

	accDRelState := NewRelativeAccount([64]byte{'3'},[64]byte{},50,true,[crypto.COMM_KEY_LENGTH]byte{},nil,nil)
	accDRelState.TxCnt = 1
	accDRelState.StakingBlockHeight = 15
	relAccounts2[[64]byte{'3'}] = &accDRelState

	accERelState := NewRelativeAccount([64]byte{'4'},[64]byte{},-90,true,[crypto.COMM_KEY_LENGTH]byte{},nil,nil)
	accERelState.TxCnt = 40
	accERelState.StakingBlockHeight = 11
	relAccounts3[[64]byte{'2'}] = &accERelState

	accFRelState := NewRelativeAccount([64]byte{'5'},[64]byte{},30,true,[crypto.COMM_KEY_LENGTH]byte{},nil,nil)
	accFRelState.TxCnt = 4
	accFRelState.StakingBlockHeight = 11
	relAccounts3[[64]byte{'3'}] = &accFRelState

	var stateTransision1 = NewStateTransition(relAccounts1,10,3,[32]byte{'1'})
	var stateTransision2 = NewStateTransition(relAccounts2,20,4,[32]byte{'2'})
	var stateTransision3 = NewStateTransition(relAccounts3,30,1,[32]byte{'3'})

	sampleStash.Set(stateTransision1.HashTransition(),stateTransision1)
	sampleStash.Set(stateTransision2.HashTransition(),stateTransision2)
	sampleStash.Set(stateTransision3.HashTransition(),stateTransision3)

	if !reflect.DeepEqual(3, len(sampleStash.M)) && !reflect.DeepEqual(3, len(sampleStash.Keys)){
		t.Errorf("Stash size does not equal 3")
	}

	var duplicateHash = stateTransision1.HashTransition()
	stateTransision4 := NewStateTransition(relAccounts3,40,5,[32]byte{'4'})
	sampleStash.Set(duplicateHash,stateTransision4)

	if !reflect.DeepEqual(3, len(sampleStash.M)) && !reflect.DeepEqual(3, len(sampleStash.Keys)){
		t.Errorf("Stash included a duplicate state transition")
	}


	//Test CheckforHeight
	var expectedHeightCountForHeight10 = 1
	var retrievedHeightCounetForHeight10 = CheckForHeightStateTransition(sampleStash,10)
	if !reflect.DeepEqual(expectedHeightCountForHeight10, retrievedHeightCounetForHeight10){
		t.Errorf("Error check for height 10 in state stash - retrieved: %d",retrievedHeightCounetForHeight10)
	}

	var expectedHeightCountForHeight20 = 1
	var retrievedHeightCounetForHeight20 = CheckForHeightStateTransition(sampleStash,20)
	if !reflect.DeepEqual(expectedHeightCountForHeight20, retrievedHeightCounetForHeight20){
		t.Errorf("Error check for height 20 in state stash - retrieved: %d",retrievedHeightCounetForHeight20)
	}

	var expectedHeightCountForHeight30 = 1
	var retrievedHeightCounetForHeight30 = CheckForHeightStateTransition(sampleStash,30)
	if !reflect.DeepEqual(expectedHeightCountForHeight30, retrievedHeightCounetForHeight30){
		t.Errorf("Error check for height 30 in state stash - retrieved: %d",retrievedHeightCounetForHeight30)
	}


	//Check retrieving state transitions for height
	var expectedStateTransitionCountForHeight10 = 1
	var retrievedStateTransitionCountForHeight10 = ReturnStateTransitionForHeight(sampleStash,10)
	if !reflect.DeepEqual(len(retrievedStateTransitionCountForHeight10), expectedStateTransitionCountForHeight10){
		t.Errorf("Error retrieval of state transition for height 10 in state stash - retrieved: %d",retrievedStateTransitionCountForHeight10)
	}

	var expectedStateTransitionCountForHeight20 = 1
	var retrievedStateTransitionCountForHeight20 = ReturnStateTransitionForHeight(sampleStash,20)
	if !reflect.DeepEqual(len(retrievedStateTransitionCountForHeight20), expectedStateTransitionCountForHeight20){
		t.Errorf("Error retrieval of state transition for height 20 in state stash - retrieved: %d",retrievedStateTransitionCountForHeight20)
	}

	var expectedStateTransitionCountForHeight30 = 1
	var retrievedStateTransitionCountForHeight30 = ReturnStateTransitionForHeight(sampleStash,30)
	if !reflect.DeepEqual(len(retrievedStateTransitionCountForHeight30), expectedStateTransitionCountForHeight30){
		t.Errorf("Error retrieval of state transition for height 20 in state stash - retrieved: %d",retrievedStateTransitionCountForHeight30)
	}
}

//func TestCheckHeightCharacteristics(t *testing.T) {
//	var sampleStash= NewBlockStash()
//
//	var blockHash= [32]byte{'0'}
//	var block= NewBlock([32]byte{'0'}, 0)
//
//	/*Fill the stash with 50 blocks*/
//	for i := 0; i < 40; i++ {
//		blockHash = [32]byte{byte(i)}
//		block = NewBlock(blockHash, uint32(i))
//		block.Hash = blockHash
//
//		sampleStash.Set(blockHash, block)
//	}
//
//	var expectedHeightCount = 1
//	var retrievedHeightCountForHeight5 = CheckForHeight(sampleStash,5)
//	var retrievedHeightCountForHeight10 = CheckForHeight(sampleStash,10)
//	var retrievedHeightCountForHeight15 = CheckForHeight(sampleStash,15)
//	var retrievedHeightCountForHeight20 = CheckForHeight(sampleStash,20)
//	var retrievedHeightCountForHeight25 = CheckForHeight(sampleStash,25)
//	var retrievedHeightCountForHeight30 = CheckForHeight(sampleStash,30)
//	var retrievedHeightCountForHeight35 = CheckForHeight(sampleStash,35)
//
//	if(expectedHeightCount != retrievedHeightCountForHeight5 || expectedHeightCount != retrievedHeightCountForHeight10 ||
//		expectedHeightCount != retrievedHeightCountForHeight15 || expectedHeightCount != retrievedHeightCountForHeight20 ||
//		expectedHeightCount != retrievedHeightCountForHeight25 || expectedHeightCount != retrievedHeightCountForHeight30 ||
//		expectedHeightCount != retrievedHeightCountForHeight35){
//		t.Errorf("Error counting entries with distinct heights")
//	}
//
//	/*Add three more entries with height 10*/
//	for i := 40; i < 42; i++ {
//		blockHash = [32]byte{byte(i)}
//		block = NewBlock(blockHash, 10)
//		block.Hash = blockHash
//
//		sampleStash.Set(blockHash, block)
//	}
//
//	var expectedHeightCountForHeight10 = 3
//	retrievedHeightCountForHeight10 = CheckForHeight(sampleStash,10)
//
//	if(expectedHeightCountForHeight10 != retrievedHeightCountForHeight10){
//		t.Errorf("Error counting entries height 10")
//	}
//
//	var expectedHashesForHeight10 = [][32]byte{[32]byte{byte(10)},[32]byte{byte(40)},[32]byte{byte(41)}}
//	var retrievedHashesForHeight10 = ReturnHashesForHeight(sampleStash,10)
//
//	var hashesCorrect = true
//
//	for _, expectedHash := range expectedHashesForHeight10 {
//		if(HashSliceContains(retrievedHashesForHeight10,expectedHash) == false){
//			hashesCorrect = false
//		}
//	}
//
//	for _, retrievedHash := range retrievedHashesForHeight10 {
//		if(HashSliceContains(expectedHashesForHeight10,retrievedHash) == false){
//			hashesCorrect = false
//		}
//	}
//
//	if(hashesCorrect == false){
//		t.Errorf("Expected and Retrieved Hashes do not match")
//	}
//}


func TestStateStashSetWhenSizeOver50Entries(t *testing.T) {
	var sampleStash = NewStateStash()

	var relAccounts1 = make(map[[64]byte]*RelativeAccount)

	//Account information in the relative state of transition 1
	accARelState := NewRelativeAccount([64]byte{'0'},[64]byte{},-10,true,[crypto.COMM_KEY_LENGTH]byte{},nil,nil)
	accARelState.TxCnt = 9
	accARelState.StakingBlockHeight = 10
	relAccounts1[[64]byte{'0'}] = &accARelState


	accBRelState := NewRelativeAccount([64]byte{'1'},[64]byte{},10,true,[crypto.COMM_KEY_LENGTH]byte{},nil,nil)
	accBRelState.TxCnt = 9
	accBRelState.StakingBlockHeight = 10
	relAccounts1[[64]byte{'1'}] = &accBRelState

	/*Fill the stash with 50 state transitions*/
	for i := 0; i < 50; i++ {
		var stateTransision1 = NewStateTransition(relAccounts1,i,3,[32]byte{'1'})

		sampleStash.Set(stateTransision1.HashTransition(),stateTransision1)
	}

	if !reflect.DeepEqual(50, len(sampleStash.M)) && !reflect.DeepEqual(50, len(sampleStash.Keys)){
		t.Errorf("Error in filling the stash: Length should be: %d - Lenght is actually: %d",50,len(sampleStash.M))
	}

	//Keep track of first entry in the stash, this one should be deleted
	firstHash,firstST := ReturnStateTransitionForPosition(sampleStash,0)
	if !reflect.DeepEqual(0, firstST.Height) && !reflect.DeepEqual([32]byte{'0'}, firstHash){
		t.Errorf("Error retrieving the first entry of the state stash")
	}

	secondHash,secondST := ReturnStateTransitionForPosition(sampleStash,1)
	if !reflect.DeepEqual(1, secondST.Height) && !reflect.DeepEqual([32]byte{'1'}, secondHash){
		t.Errorf("Error retrieving the second entry of the state stash")
	}

	thirdHash,thirdST := ReturnStateTransitionForPosition(sampleStash,2)
	if !reflect.DeepEqual(2, thirdST.Height) && !reflect.DeepEqual([32]byte{'2'}, thirdHash){
		t.Errorf("Error retrieving the third entry of the statestash")
	}

	outofboundHash,outofboundBlock := ReturnStateTransitionForPosition(sampleStash,50)
	if(outofboundHash != [32]byte{} && outofboundBlock != nil){
		t.Errorf("Error expected out of bound exception")
	}

	/*Add another block to the stash and thus, delete first entry*/
	var stateTransision1 = NewStateTransition(relAccounts1,50,3,[32]byte{'1'})
	sampleStash.Set(stateTransision1.HashTransition(),stateTransision1)

	firstHash,firstST = ReturnStateTransitionForPosition(sampleStash,0)
	if !reflect.DeepEqual(1, firstST.Height){
		t.Errorf("Error deleting first entryof the stash")
	}

	var stateTransision2 = NewStateTransition(relAccounts1,51,3,[32]byte{'1'})
	sampleStash.Set(stateTransision1.HashTransition(),stateTransision2)

	var stateTransision3 = NewStateTransition(relAccounts1,52,3,[32]byte{'1'})
	sampleStash.Set(stateTransision1.HashTransition(),stateTransision3)

	var stateTransision4 = NewStateTransition(relAccounts1,53,3,[32]byte{'1'})
	sampleStash.Set(stateTransision1.HashTransition(),stateTransision4)

	var stateTransision5 = NewStateTransition(relAccounts1,54,3,[32]byte{'1'})
	sampleStash.Set(stateTransision1.HashTransition(),stateTransision5)

	if !reflect.DeepEqual(50, len(sampleStash.M)) && !reflect.DeepEqual(50, len(sampleStash.Keys)){
		t.Errorf("Error keeping stash size at 50")
	}
}

//func HashSliceContains(slice [][32]byte, hash [32]byte) bool {
//	for _, a := range slice {
//		if a == hash {
//			return true
//		}
//	}
//	return false
//}