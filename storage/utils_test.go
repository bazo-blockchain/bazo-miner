package storage

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"math/big"
	"testing"
)

func TestSerializeHashContent(t *testing.T) {
	var data []byte
	pubKeyInt, _ := new(big.Int).SetString(PubA1+PubA2, 16)
	copy(data, pubKeyInt.Bytes())

	hash := protocol.SerializeHashContent(data)
	base16String := fmt.Sprintf("%x", hash)

	if base16String != "075783ca932e234acfabbe9d989c35b59c87495a77745bf79e6b704549af2cfa" {
		t.Errorf("Error serializing: %x != %v\n", hash, "075783ca932e234acfabbe9d989c35b59c87495a77745bf79e6b704549af2cfa")
	}
}

func TestGetAccount(t *testing.T) {
	accAHash := protocol.SerializeHashContent(accA.Address)

	acc, err := GetAccount(accAHash)

	if acc != accA && err == nil {
		t.Errorf("Error fetching account from state: %x\n", accAHash)
	}

	if acc == accB && err == nil {
		t.Errorf("Error fetching account from state: %x\n", accAHash)
	}

	var nilHash [32]byte
	acc, err = GetAccount(nilHash)

	if acc != nil || err.Error() != fmt.Sprintf("Acc (%x) not in the state.", nilHash[0:8]) {
		t.Errorf("Error fetching account from state: %x\n", nilHash)
	}
}

func TestGetRootAccount(t *testing.T) {
	rootHash := protocol.SerializeHashContent(rootAcc.Address)

	root, err := GetRootAccount(rootHash)

	if root == nil || err != nil {
		t.Errorf("Error fetching root account from state: %x\n", rootHash)
	}

	var nilHash [32]byte
	root, err = GetRootAccount(nilHash)

	if root != nil {
		t.Errorf("Error fetching account from state: %x\n", nilHash)
	}
}
