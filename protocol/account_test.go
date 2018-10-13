package protocol

import (
	"reflect"
	"testing"
)

func TestAccountSerialization(t *testing.T) {
	addTestingAccounts()

	accA.Balance = 1000
	accA.IsStaking = true
	accA.TxCnt = 5
	accA.StakingBlockHeight = 100

	var compareAcc *Account
	encodedAcc := accA.Encode()
	compareAcc = compareAcc.Decode(encodedAcc)

	if !reflect.DeepEqual(accA, compareAcc) {
		t.Error("Account encoding/decoding failed!")
	}
}
