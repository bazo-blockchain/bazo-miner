package miner

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"math/big"

	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

// This test deploys a smart contract in the first block and calls the smart contract in the second block
func TestMultipleBlocksWithContractTx(t *testing.T) {
	cleanAndPrepare()

	b := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	contract := []byte{
		35,      // CALLDATA
		0, 0, 5, // PUSH 5
		4,  // ADD
		49, // HALT
	}
	createBlockWithSingleContractDeployTx(b, contract, nil)
	finalizeBlock(b)
	if err := validateBlock(b); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	b2 := newBlock(b.Hash, [32]byte{}, [32]byte{}, 2)
	transactionData := []byte{
		0, 15,
	}
	createBlockWithSingleContractCallTx(b2, transactionData)
	finalizeBlock(b2)
	if err := validateBlock(b2); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}
}

// This test deploys a smart contract with a state variable in the first block and calls the smart contract in the second
// block which loads the state variable, alters the local variable and stores the change
func TestMultipleBlocksWithStateChangeContractTx(t *testing.T) {
	cleanAndPrepare()

	b := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	contract := []byte{
		35,    // CALLDATA
		29, 0, // SLOAD
		4,     // ADD
		27, 0, // SSTORE
		49, // HALT
	}
	createBlockWithSingleContractDeployTx(b, contract, []big.Int{*big.NewInt(2)})
	finalizeBlock(b)
	if err := validateBlock(b); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	b2 := newBlock(b.Hash, [32]byte{}, [32]byte{}, 2)
	transactionData := []byte{
		0, 15,
	}
	hash := createBlockWithSingleContractCallTx(b2, transactionData)
	finalizeBlock(b2)
	if err := validateBlock(b2); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}

	contractVariables := storage.GetAccount(hash).ContractVariables
	if !reflect.DeepEqual(contractVariables, []big.Int{*big.NewInt(17)}) {
		t.Errorf("State change not persisted, expected: [{false [17]}], is %v.", contractVariables)
	}
}

// This test is similar to the TestMultipleBlocksWithStateChangeContractTx. The difference is, that after the first state change
// transaction, a second one is called, which changes the state again.
func TestMultipleBlocksWithDoubleStateChangeContractTx(t *testing.T) {
	cleanAndPrepare()

	b := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	contract := []byte{
		35,    // CALLDATA
		29, 0, // SLOAD
		4,     // ADD
		27, 0, // SSTORE
		49, // HALT
	}
	createBlockWithSingleContractDeployTx(b, contract, []big.Int{*big.NewInt(2)})
	finalizeBlock(b)
	if err := validateBlock(b); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	b2 := newBlock(b.Hash, [32]byte{}, [32]byte{}, 2)
	transactionData := []byte{
		0, 15,
	}
	createBlockWithSingleContractCallTx(b2, transactionData)
	finalizeBlock(b2)
	if err := validateBlock(b2); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}

	b3 := newBlock(b2.Hash, [32]byte{}, [32]byte{}, 3)
	transactionData = []byte{
		0, 15,
	}
	hash := createBlockWithSingleContractCallTx(b3, transactionData)
	finalizeBlock(b3)
	if err := validateBlock(b3); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}

	contractVariables := storage.GetAccount(hash).ContractVariables
	if !reflect.DeepEqual(contractVariables, []big.Int{*big.NewInt(32)}) {
		t.Errorf("State change not persisted, expected: [{false [32]}], is %v.", contractVariables)
	}
}

func TestMultipleBlocksWithContextContractTx(t *testing.T) {
	cleanAndPrepare()

	b := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	contract := []byte{
		35, 0, 0, 1, 10, 22, 0, 10, 1, 49, 28, 0, 31, 33, 10, 22, 0, 21, 2, 24, 28, 0, 29, 0, 0, 4, 27, 0, 0, 24,
	}
	createBlockWithSingleContractDeployTx(b, contract, nil)
	finalizeBlock(b)
	if err := validateBlock(b); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	b1 := newBlock(b.Hash, [32]byte{}, [32]byte{}, 2)
	transactionData := []byte{
		0, 100, // Amount
		0, 1,
	}
	createBlockWithSingleContractCallTx(b1, transactionData)
	finalizeBlock(b1)
	if err := validateBlock(b1); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}
}

// This test deploys a smart contract in the first block and calls the smart contract in the second block
func TestMultipleBlocksWithTokenizationContractTx(t *testing.T) {
	cleanAndPrepare()

	b := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	contract := []byte{
		35, 1, 0, 0, 1, 10, 22, 0, 11, 3, 49, 28, 0, 28, 1, 31, 33, 10, 22, 0, 24, 2, 24, 28, 1, 29, 0, 0, 38, 28, 0, 4, 39, 27, 0, 0, 24,
	}
	issuer := createBlockWithSingleContractDeployTx(b, contract, nil)
	finalizeBlock(b)
	if err := validateBlock(b); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	b1 := newBlock(b.Hash, [32]byte{}, [32]byte{}, 2)
	transactionData := []byte{
		0, 100, // Amount
		31, // Length of next parameter (receiver address)
	}
	transactionData = append(transactionData, issuer[:]...) // append receiver address
	transactionData = append(transactionData, 0, 1)         // append function hash
	createBlockWithSingleContractCallTx(b1, transactionData)
	finalizeBlock(b1)
	if err := validateBlock(b1); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}
}

func createBlockWithSingleContractDeployTx(b *protocol.Block, contract []byte, contractVariables []big.Int) [32]byte {
	tx, _, _ := protocol.ConstrAccTx(0, rand.Uint64()%100+1, [64]byte{}, &RootPrivKey, contract, contractVariables)
	if err := addTx(b, tx); err == nil {
		storage.WriteOpenTx(tx)
		return tx.Issuer
	} else {
		fmt.Print(err)
		return [32]byte{}
	}
}

func createBlockWithSingleContractCallTx(b *protocol.Block, transactionData []byte) [32]byte {
	for hash := range storage.GetAllAccounts() {
		if storage.GetAccount(hash).Contract != nil {
			accAHash := protocol.SerializeHashContent(accA.Address)
			accBHash := storage.GetAccount(hash).Hash()

			tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, uint32(accA.TxCnt), accAHash, accBHash, &PrivKeyA, &multiSignPrivKeyA, transactionData)
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
	accAHash := storage.GetAccount(from).Hash()
	accBHash := storage.GetAccount(to).Hash()

	tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, uint32(accA.TxCnt), accAHash, accBHash, &PrivKeyA, &multiSignPrivKeyA, transactionData)
	if err := addTx(b, tx); err == nil {
		storage.WriteOpenTx(tx)
	} else {
		fmt.Print(err)
	}
}

func getAccountsWithContracts() []protocol.Account {
	var accounts []protocol.Account
	for hash := range storage.GetAllAccounts() {
		if storage.GetAccount(hash).Contract != nil {
			accounts = append(accounts, *storage.GetAccount(hash))
		}
	}
	return accounts
}