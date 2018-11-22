package miner

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"math/rand"
	"reflect"
	"testing"

	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/bazo-blockchain/bazo-miner/vm"
)

// This test deploys a smart contract in the first block and calls the smart contract in the second block
func TestMultipleBlocksWithContractTx(t *testing.T) {
	cleanAndPrepare()

	b := newBlock([32]byte{}, [crypto.COMM_PROOF_LENGTH]byte{}, 1)
	contract := []byte{
		35,         // CALLDATA
		0, 1, 0, 5, // PUSH 5
		4,  // ADD
		50, // HALT
	}
	createBlockWithSingleContractDeployTx(b, contract, nil)
	finalizeBlock(b)
	if err := validate(b, false); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	b2 := newBlock(b.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, 2)
	transactionData := []byte{
		1, 0, 15,
	}
	createBlockWithSingleContractCallTx(b2, transactionData)
	finalizeBlock(b2)
	if err := validate(b2, false); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}
}

// This test deploys a smart contract with a state variable in the first block and calls the smart contract in the second
// block which loads the state variable, alters the local variable and stores the change
func TestMultipleBlocksWithStateChangeContractTx(t *testing.T) {
	cleanAndPrepare()

	b := newBlock([32]byte{}, [crypto.COMM_PROOF_LENGTH]byte{}, 1)
	contract := []byte{
		35,    // CALLDATA
		29, 0, // SLOAD
		4,     // ADD
		27, 0, // SSTORE
		50, // HALT
	}
	createBlockWithSingleContractDeployTx(b, contract, []protocol.ByteArray{[]byte{0, 2}})
	finalizeBlock(b)
	if err := validate(b, false); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	b2 := newBlock(b.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, 2)
	transactionData := []byte{
		1, 0, 15,
	}
	hash := createBlockWithSingleContractCallTx(b2, transactionData)
	finalizeBlock(b2)
	if err := validate(b2, false); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}

	acc, _ := storage.GetAccount(hash)
	contractVariables := acc.ContractVariables
	expected := []protocol.ByteArray{[]byte{0, 17}}
	if !reflect.DeepEqual(contractVariables, expected) {
		t.Errorf("State change not persisted, expected: '%v', is '%v'.", expected, contractVariables)
	}
}

// This test is similar to the TestMultipleBlocksWithStateChangeContractTx. The difference is, that after the first state change
// transaction, a second one is called, which changes the state again.
func TestMultipleBlocksWithDoubleStateChangeContractTx(t *testing.T) {
	cleanAndPrepare()

	b := newBlock([32]byte{}, [crypto.COMM_PROOF_LENGTH]byte{}, 1)
	contract := []byte{
		35,    // CALLDATA
		29, 0, // SLOAD
		4,     // ADD
		27, 0, // SSTORE
		50, // HALT
	}
	createBlockWithSingleContractDeployTx(b, contract, []protocol.ByteArray{[]byte{0, 2}})
	finalizeBlock(b)
	if err := validate(b, false); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	b2 := newBlock(b.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, 2)
	transactionData := []byte{
		1, 0, 15,
	}
	createBlockWithSingleContractCallTx(b2, transactionData)
	finalizeBlock(b2)
	if err := validate(b2, false); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}

	b3 := newBlock(b2.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, 3)
	transactionData = []byte{
		1, 0, 15,
	}
	hash := createBlockWithSingleContractCallTx(b3, transactionData)
	finalizeBlock(b3)
	if err := validate(b3, false); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}

	acc, _ := storage.GetAccount(hash)
	contractVariables := acc.ContractVariables
	expected := []protocol.ByteArray{[]byte{0, 32}}
	if !reflect.DeepEqual(contractVariables, expected) {
		t.Errorf("State change not persisted, expected: '%v', is %v.", expected, contractVariables)
	}
}

func TestMultipleBlocksWithContextContractTx(t *testing.T) {
	cleanAndPrepare()

	b := newBlock([32]byte{}, [crypto.COMM_PROOF_LENGTH]byte{}, 1)
	contract := []byte{
		35, 0, 0, 1, 10, 22, 0, 10, 1, 50, 28, 0, 31, 33, 10, 22, 0, 21, 2, 24, 28, 0, 29, 0, 0, 4, 27, 0, 0, 24,
	}
	createBlockWithSingleContractDeployTx(b, contract, nil)
	finalizeBlock(b)
	if err := validate(b, false); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	b1 := newBlock(b.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, 2)
	transactionData := []byte{
		0, 100, // Amount
		0, 1,
	}
	createBlockWithSingleContractCallTx(b1, transactionData)
	finalizeBlock(b1)
	if err := validate(b1, false); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}
}

// This test deploys a smart contract in the first block and calls the smart contract in the second block
func TestMultipleBlocksWithTokenizationContractTx(t *testing.T) {
	cleanAndPrepare()

	b := newBlock([32]byte{}, [crypto.COMM_PROOF_LENGTH]byte{}, 1)
	contract := []byte{
		35, 1, 0, 0, 1, 10, 22, 0, 11, 3, 50, 28, 1, 28, 0, 29, 1, 33, 10, 22, 0, 24, 2, 24, 28, 1, 28, 0, 1, 29, 2, 37, 22, 0, 46, 2, 28, 1, 28, 0, 29, 2, 38, 27, 2, 50, 28, 1, 29, 2, 39, 28, 0, 4, 28, 1, 29, 2, 40, 27, 2, 50,
	}

	contractVariables := make([]protocol.ByteArray, 3)
	receiver := []byte{0x00, 0x2b}
	contractVariables[0] = receiver

	accAHash := protocol.SerializeHashContent(accA.Address)
	var minter = accAHash[:]
	contractVariables[1] = minter

	m := vm.NewMap()
	m.Append(receiver, []byte{0x00, 0x01})
	contractVariables[2] = []byte(m)

	createBlockWithSingleContractDeployTx(b, contract, contractVariables)
	finalizeBlock(b)
	if err := validate(b, false); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	b1 := newBlock(b.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, 2)
	transactionData := []byte{
		1, 0, 100, // Amount
		1, receiver[0], receiver[1], // receiver address
		1, 0, 1, // function Hash
	}
	hash := createBlockWithSingleContractCallTx(b1, transactionData)
	finalizeBlock(b1)
	if err := validate(b1, false); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}

	acc, _ := storage.GetAccount(hash)
	m, err := vm.MapFromByteArray(acc.ContractVariables[2])
	if err != nil {
		t.Errorf(err.Error())
	}

	tmp, err := m.GetVal(receiver)
	if err != nil {
		t.Errorf(err.Error())
	}

	actual := uint64(tmp[1])
	expected := uint64(101)
	if expected != actual {
		t.Errorf("State change not persisted, expected: '%v', but is: '%v'", expected, actual)
	}
}

func TestMultipleBlocksWithTokenizationContractTxWhichAddsKey(t *testing.T) {
	cleanAndPrepare()

	b := newBlock([32]byte{}, [crypto.COMM_PROOF_LENGTH]byte{}, 1)
	contract := []byte{
		35, 1, 0, 0, 1, 10, 22, 0, 11, 3, 50, 28, 1, 28, 0, 29, 1, 33, 10, 22, 0, 24, 2, 24, 28, 1, 28, 0, 1, 29, 2, 37, 22, 0, 46, 2, 28, 1, 28, 0, 29, 2, 38, 27, 2, 50, 28, 1, 29, 2, 39, 28, 0, 4, 28, 1, 29, 2, 40, 27, 2, 50,
	}

	contractVariables := make([]protocol.ByteArray, 3)
	receiver := []byte{0x00, 0x2b}
	contractVariables[0] = receiver

	accAHash := protocol.SerializeHashContent(accA.Address)
	var minter = accAHash[:]
	contractVariables[1] = minter

	m := vm.NewMap()
	//m.Append(receiver, []byte{0x00, 0x01})
	contractVariables[2] = []byte(m)

	createBlockWithSingleContractDeployTx(b, contract, contractVariables)
	finalizeBlock(b)
	if err := validate(b, false); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	b1 := newBlock(b.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, 2)
	transactionData := []byte{
		1, 0, 100, // Amount
		1, receiver[0], receiver[1], // receiver address
		1, 0, 1, // function Hash
	}
	hash := createBlockWithSingleContractCallTx(b1, transactionData)
	finalizeBlock(b1)
	if err := validate(b1, false); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}

	acc, _ := storage.GetAccount(hash)
	m, err := vm.MapFromByteArray(acc.ContractVariables[2])
	if err != nil {
		t.Errorf(err.Error())
	}

	tmp, err := m.GetVal(receiver)
	if err != nil {
		t.Errorf(err.Error())
	}

	actual := uint64(tmp[1])
	expected := uint64(100)
	if expected != actual {
		t.Errorf("State change not persisted, expected: '%v', but is: '%v'", expected, actual)
	}
}

func createBlockWithSingleContractDeployTx(b *protocol.Block, contract []byte, contractVariables []protocol.ByteArray) [32]byte {
	tx, _, _ := protocol.ConstrAccTx(0, 1000000, [64]byte{}, PrivKeyRoot, contract, contractVariables)
	if err := addTx(b, tx); err == nil {
		storage.WriteOpenTx(tx)
		return tx.Issuer
	} else {
		fmt.Print(err)
		return [32]byte{}
	}
}

func createBlockWithSingleContractCallTx(b *protocol.Block, transactionData []byte) [32]byte {
	for hash := range storage.State {
		acc, _ := storage.GetAccount(hash)
		if acc.Contract != nil {
			accAHash := protocol.SerializeHashContent(accA.Address)
			accBHash := acc.Hash()

			tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100+1, 100000, uint32(accA.TxCnt), accAHash, accBHash, PrivKeyAccA, PrivKeyMultiSig, transactionData)
			if err := addTx(b, tx); err == nil {
				storage.WriteOpenTx(tx)
			} else {
				fmt.Print(err)
			}
			return accBHash
		}
	}
	return [32]byte{}
}

func createBlockWithSingleContractCallTxDefined(b *protocol.Block, transactionData []byte, from [32]byte, to [32]byte) {
	accA, _ := storage.GetAccount(from)
	accB, _ := storage.GetAccount(to)

	tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, uint32(accA.TxCnt), accA.Hash(), accB.Hash(), PrivKeyAccA, PrivKeyMultiSig, transactionData)
	if err := addTx(b, tx); err == nil {
		storage.WriteOpenTx(tx)
	} else {
		fmt.Print(err)
	}
}

func getAccountsWithContracts() []protocol.Account {
	var accounts []protocol.Account
	for hash := range storage.State {
		acc, _ := storage.GetAccount(hash)
		if acc.Contract != nil {
			accounts = append(accounts, *acc)
		}
	}
	return accounts
}
