package miner

import (
	"testing"
	rand "math/rand"
	crand "crypto/rand"
	"time"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"crypto/ecdsa"
	"crypto/elliptic"
)

/**
 TODO: strategy to decide where and when to start the consolidation
 TODO: strategy to decide how many blocks should be included (ignore latest X blocks)
Should X be part of the configuration?

 TODO: who does the consolidation?
     Anybody should be able to create a consolidationTx


TODO: Check whether the miner consolidated it correctly so we cannot prune it directly
TODO: Consolidation between genesis/previous consolidation and latest X blocks
 Enforce rule that a consolidationTx:
   - check if there are some tx that have been left out, if yes then reject the block.
 */
var InitialBalance uint64 = 100000
/**
 * Steps for basic test
 *
 * Create a chain with blocks filled with transactions from address A to address B
 * - keep track of the amount
 * - go through all the blocks of the chain and sum the results
 * - create a new consolidation block where balance of B is the sum
 */



func getNewAddr() (hash [32]byte){
	newKey, _ := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	newAccPub1, newAccPub2 := newKey.PublicKey.X.Bytes(), newKey.PublicKey.Y.Bytes()
	var pubKey [64]byte
	copy(pubKey[0:32], newAccPub1)
	copy(pubKey[32:64], newAccPub2)
	address := storage.GetAddressFromPubKey(&newKey.PublicKey)
	return storage.SerializeHashContent(address)
}

func getTestState() (protocol.StateAccounts){
	teststate := make(protocol.StateAccounts)
	addr := getNewAddr()
	fmt.Printf("Test Addr used for consolidation: %v", addr)
	consAcc := new(protocol.ConsolidatedAccount)
	consAcc.Account = addr
	consAcc.Balance = 1000
	consAcc.Staking = false
	teststate[addr] = consAcc
	return teststate
}

func createBlock(t *testing.T, b *protocol.Block) ([][32]byte, [][32]byte, [][32]byte, [][32]byte) {

	var testSize uint32
	testSize = 100

	var hashFundsSlice [][32]byte
	var hashAccSlice [][32]byte
	var hashConfigSlice [][32]byte
	var hashStakeSlice [][32]byte

	//in order to create valid funds transactions we need to know the tx count of acc A

	rand := rand.New(rand.NewSource(time.Now().Unix()))
	loopMax := int(rand.Uint32()%testSize) + 1
	loopMax += int(accA.TxCnt)
	for cnt := int(accA.TxCnt); cnt < loopMax; cnt++ {
		accAHash := protocol.SerializeHashContent(accA.Address)
		accBHash := protocol.SerializeHashContent(accB.Address)
		accA.Balance = InitialBalance
		accB.Balance= InitialBalance
		tx, txErr := protocol.ConstrFundsTx(0x01, 50, 1, uint32(cnt), accAHash, accBHash, &PrivKeyA, &multiSignPrivKeyA)
		if txErr != nil {
			t.Error(txErr)
		}
		if err := addTx(b, tx); err == nil {
			//Might be that we generated a block that was already generated before
			if storage.ReadOpenTx(tx.Hash()) != nil || storage.ReadClosedTx(tx.Hash()) != nil {
				continue
			}
			hashFundsSlice = append(hashFundsSlice, tx.Hash())
			storage.WriteOpenTx(tx)
		} else {
			fmt.Print(err)
		}
	}

	return hashFundsSlice, hashAccSlice, hashConfigSlice, hashStakeSlice
}

func createTestChain(t *testing.T)([]*protocol.Block) {
	var blockList []*protocol.Block
	var numberOfTestBlocks = 2
	testState := getTestState()
	prevHash := [32]byte{}

	// Create blocks filled with random transactions, finalize (PoW etc.) and validate (state change)
	for cnt := int(accA.TxCnt); cnt < numberOfTestBlocks; cnt++ {
		b := newBlock(prevHash, [32]byte{}, [32]byte{}, uint32(cnt))
		createBlock(t, b)

		consTx, err := protocol.ConstrConsolidationTx(0, testState, prevHash)
		if err != nil {
			t.Errorf("Could not create test consolidationTx: %v\n", err)
		}
		if cnt == 1 {
			if err := addTx(b, consTx); err == nil {
				//Might be that we generated a block that was already generated before
				if storage.ReadOpenTx(consTx.Hash()) != nil || storage.ReadClosedTx(consTx.Hash()) != nil {
					continue
				}
				storage.WriteOpenTx(consTx)
			} else {
				fmt.Print(err)
			}
		}
		// Add consolidation tx
		if err := finalizeBlock(b); err != nil {
			t.Errorf("Could not finalize block: %v\n", err)
		}
		if err := validateBlock(b); err != nil {
			t.Errorf("Could not validate block: %v\n", err)
		}
		prevHash = b.Hash
		blockList = append(blockList, b)
	}
	return blockList
}



func TestBasicConsolidationTx(t *testing.T) {
	cleanAndPrepare()
	chain := createTestChain(t)

	consTx, err := GetConsolidationTxFromChain(chain)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(consTx)
}