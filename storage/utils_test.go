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

  if fmt.Sprintf("%x", hash) != "ca4510738395af1429224dd785675309c344b2b549632e20275c69b15ed1d210" {
		t.Errorf("Error serializing: %x != %v\n", hash, "ca4510738395af1429224dd785675309c344b2b549632e20275c69b15ed1d210")
	}
}

func TestGetAccount(t *testing.T) {
	acc, err := ReadAccount(accA.Address)

	if acc != accA && err == nil {
		t.Errorf("Error fetching account from state: %x\n", accA.Address)
	}

	if acc == accB && err == nil {
		t.Errorf("Error fetching account from state: %x\n", accA.Address)
	}

	var nilAddress [64]byte
	acc, err = ReadAccount(nilAddress)

	if acc != nil || err.Error() != fmt.Sprintf("Acc (%x) not in the state.", nilAddress[0:8]) {
		t.Errorf("Error fetching account from state: %x\n", nilAddress)
	}
}

func TestGetRootAccount(t *testing.T) {
	root, err := ReadRootAccount(rootAcc.Address)

	if root == nil || err != nil {
		t.Errorf("Error fetching root account from state: %x\n", rootAcc.Address)
	}

	var nilAddress [64]byte
	root, err = ReadRootAccount(nilAddress)

	if root != nil {
		t.Errorf("Error fetching account from state: %x\n", nilAddress)
	}
}
