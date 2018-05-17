package miner

import (
	"fmt"
	"math/rand"
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
		33,      // CALLDATA
		0, 0, 5, // PUSH 5
		4,  // ADD
		45, // HALT
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
		33,    // CALLDATA
		28, 0, // SLOAD
		4,     // ADD
		26, 0, // SSTORE
		46, // HALT
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
}

func createBlockWithSingleContractDeployTx(b *protocol.Block, contract []byte, contractVariables []big.Int) {
	tx, _, _ := protocol.ConstrAccTx(0, rand.Uint64()%100+1, [64]byte{}, &RootPrivKey, contract, contractVariables)

	if err := addTx(b, tx); err == nil {
		storage.WriteOpenTx(tx)
	} else {
		fmt.Print(err)
	}
}

func createBlockWithSingleContractCallTx(b *protocol.Block, transactionData []byte) {
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
		}
	}
}
