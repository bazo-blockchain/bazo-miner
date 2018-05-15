package miner

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

// This test deploys a smart contract in the first block and calls the smart contract in the second block
func TestMultipleBlocksWithSingleContract(t *testing.T) {
	cleanAndPrepare()

	b := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	createBlockWithSingleContractDeployTx(b)
	finalizeBlock(b)
	if err := validateBlock(b); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	b2 := newBlock(b.Hash, [32]byte{}, [32]byte{}, 2)
	createBlockWithSingleContractCallTx(b2)
	finalizeBlock(b2)
	if err := validateBlock(b2); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}
}

func createBlockWithSingleContractDeployTx(b *protocol.Block) {
	contract := []byte{
		33,      // CALLDATA
		0, 0, 5, // PUSH 5
		4,  // ADD
		45, // HALT
	}

	tx, _, _ := protocol.ConstrAccTx(0, rand.Uint64()%100+1, [64]byte{}, &RootPrivKey, contract, nil)

	if err := addTx(b, tx); err == nil {
		storage.WriteOpenTx(tx)
	} else {
		fmt.Print(err)
	}
}

func createBlockWithSingleContractCallTx(b *protocol.Block) {
	for hash := range storage.GetAllAccounts() {
		if storage.GetAccount(hash).Contract != nil {
			accAHash := protocol.SerializeHashContent(accA.Address)
			accBHash := storage.GetAccount(hash).Hash()

			transactionData := []byte{
				0, 15,
			}

			tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100+1, rand.Uint64()%100+1, uint32(accA.TxCnt), accAHash, accBHash, &PrivKeyA, &multiSignPrivKeyA, transactionData)
			if err := addTx(b, tx); err == nil {
				storage.WriteOpenTx(tx)
			} else {
				fmt.Print(err)
			}
		}
	}
}
