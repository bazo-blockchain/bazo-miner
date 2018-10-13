package protocol

import (
	"reflect"
	"testing"
)

func TestAccountSerialization(t *testing.T) {
	var compareAcc *Account
	encodedAcc := accA.Encode()
	compareAcc = compareAcc.Decode(encodedAcc)

	if !reflect.DeepEqual(accA, compareAcc) {
		t.Error("Account Serialization failed!")
	}
}
