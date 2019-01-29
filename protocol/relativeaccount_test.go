package protocol

import (
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"reflect"
	"testing"
)

func TestStateTransition(t *testing.T) {
	var stateRelative = make(map[[64]byte]*RelativeAccount)

	//Account information in the relative state
	accARelState := NewRelativeAccount([64]byte{'0'},[64]byte{},-10,true,[crypto.COMM_KEY_LENGTH]byte{},nil,nil)
	accARelState.TxCnt = 9
	accARelState.StakingBlockHeight = 10
	stateRelative[[64]byte{'0'}] = &accARelState

	accBRelState := NewRelativeAccount([64]byte{'1'},[64]byte{},10,true,[crypto.COMM_KEY_LENGTH]byte{},nil,nil)
	accBRelState.TxCnt = 9
	accBRelState.StakingBlockHeight = 10
	stateRelative[[64]byte{'1'}] = &accBRelState

	accCRelState := NewRelativeAccount([64]byte{'2'},[64]byte{},0,true,[crypto.COMM_KEY_LENGTH]byte{},nil,nil)
	accCRelState.TxCnt = 9
	accCRelState.StakingBlockHeight = 10
	stateRelative[[64]byte{'2'}] = &accCRelState

	accDRelState := NewRelativeAccount([64]byte{'3'},[64]byte{},50,true,[crypto.COMM_KEY_LENGTH]byte{},nil,nil)
	accDRelState.TxCnt = 1
	accDRelState.StakingBlockHeight = 15
	stateRelative[[64]byte{'3'}] = &accDRelState

	var stateTransition = NewStateTransition(stateRelative,10,3)

	var compareTransition *StateTransition
	encodedAcc := stateTransition.EncodeTransition()
	compareTransition = compareTransition.DecodeTransition(encodedAcc)

	if !reflect.DeepEqual(stateTransition.Height, compareTransition.Height) {
		t.Error("State Transition encoding/decoding failed: Height does not match!")
	}

	if !reflect.DeepEqual(stateTransition.ShardID, compareTransition.ShardID) {
		t.Error("State Transition encoding/decoding failed: ShardID does not match!")
	}

	for k, _ := range stateTransition.RelativeStateChange {
		if _, ok := compareTransition.RelativeStateChange[k]; !ok {
			t.Errorf("account not existing in serialized state")
		} else {
			var accStateTransition = stateTransition.RelativeStateChange[k]
			var accCompareTransition = compareTransition.RelativeStateChange[k]

			if !reflect.DeepEqual(accStateTransition, accCompareTransition){
				t.Errorf("expected and retrieved relative account information does not match")
			}
		}
	}
}
