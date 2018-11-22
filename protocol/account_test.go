package protocol

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestAccountCreation(t *testing.T) {
	createdAcc := NewAccount(accA.Address, accA.Issuer, accA.Balance, accA.IsStaking, accA.CommitmentKey, accA.Contract, accA.ContractVariables)

	if !reflect.DeepEqual(createdAcc.Address, accA.Address) {
		t.Errorf("Address does not match the given one: %x vs. %x", createdAcc.Address, accA.Address)
	}

	if !reflect.DeepEqual(createdAcc.Issuer, accA.Issuer) {
		t.Errorf("Issuer does not match the given one: %x vs. %x", createdAcc.Issuer, accA.Issuer)
	}

	if !reflect.DeepEqual(createdAcc.Balance, accA.Balance) {
		t.Errorf("Balance does not match the given one: %v vs. %v", createdAcc.Balance, accA.Balance)
	}

	if !reflect.DeepEqual(createdAcc.IsStaking, accA.IsStaking) {
		t.Errorf("IsStaking does not match the given one: %v vs. %v", createdAcc.IsStaking, accA.IsStaking)
	}

	if !reflect.DeepEqual(createdAcc.CommitmentKey, accA.CommitmentKey) {
		t.Errorf("CommitmentKey does not match the given one: %x vs. %x", createdAcc.CommitmentKey, accA.CommitmentKey)
	}

	if !reflect.DeepEqual(createdAcc.Contract, accA.Contract) {
		t.Errorf("Contract does not match the given one: %x vs. %x", createdAcc.Contract, accA.Contract)
	}

	if !reflect.DeepEqual(createdAcc.ContractVariables, accA.ContractVariables) {
		t.Errorf("ContractVariables does not match the given one: %x vs. %x", createdAcc.ContractVariables, accA.ContractVariables)
	}
}

func TestAccountHash(t *testing.T) {
	var address [64]byte
	rand.Read(address[:])

	hash1 := accA.Hash()

	if !reflect.DeepEqual(hash1, accA.Hash()) {
		t.Errorf("Account hashing failed!")
	}

	accA.Address = address
	hash2 :=accA.Hash()

	if reflect.DeepEqual(hash1, hash2) {
		t.Errorf("Account hashing failed!")
	}
}

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
