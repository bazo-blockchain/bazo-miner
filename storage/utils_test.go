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

  if fmt.Sprintf("%x", hash) != "075783ca932e234acfabbe9d989c35b59c87495a77745bf79e6b704549af2cfa" {
		t.Errorf("Error serializing: %x != %v\n", hash, "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a")
	}
}

func TestGetAccount(t *testing.T) {
	acc, err := GetAccount(accA.Address)

	if acc != accA && err == nil {
		t.Errorf("Error fetching account from state: %x\n", accA.Address)
	}

	if acc == accB && err == nil {
		t.Errorf("Error fetching account from state: %x\n", accA.Address)
	}

	var nilAddress [64]byte
	acc, err = GetAccount(nilAddress)

	if acc != nil || err.Error() != fmt.Sprintf("Acc (%x) not in the state.", nilAddress[0:8]) {
		t.Errorf("Error fetching account from state: %x\n", nilAddress)
	}
}

func TestGetRootAccount(t *testing.T) {
	root, err := GetRootAccount(rootAcc.Address)

	if root == nil || err != nil {
		t.Errorf("Error fetching root account from state: %x\n", rootAcc.Address)
	}

	var nilAddress [64]byte
	root, err = GetRootAccount(nilAddress)

	if root != nil {
		t.Errorf("Error fetching account from state: %x\n", nilAddress)
	}
}
