package miner

import (
	"testing"
	"math/rand"
	"time"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/p2p"
)

/**
 TODO: strategy to decide where and when to start the consolidation
 TODO: strategy to decide how many blocks should be included
 TODO: who does the consolidation?
     Anybody should be able to create a consolidationTx

 TODO: block or simple transaction?
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
	prevHash := [32]byte{}

	// Create blocks filled with random transactions, finalize (PoW etc.) and validate (state change)
	for cnt := int(accA.TxCnt); cnt < numberOfTestBlocks; cnt++ {
		b := newBlock(prevHash, [32]byte{}, [32]byte{}, uint32(cnt))
		createBlock(t, b)
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


func reqTx(txType uint8, txHash [32]byte) (tx protocol.Transaction) {
	if conn := p2p.Connect(storage.BOOTSTRAP_SERVER); conn != nil {

		packet := p2p.BuildPacket(txType, txHash[:])
		conn.Write(packet)

		header, payload, err := p2p.RcvData(conn)
		if err != nil {
			logger.Printf("Requesting tx failed.")
			return
		}

		switch header.TypeID {
		case p2p.ACCTX_RES:
			var accTx *protocol.AccTx
			accTx = accTx.Decode(payload)
			tx = accTx
		case p2p.CONFIGTX_RES:
			var configTx *protocol.ConfigTx
			configTx = configTx.Decode(payload)
			tx = configTx
		case p2p.FUNDSTX_RES:
			var fundsTx *protocol.FundsTx
			fundsTx = fundsTx.Decode(payload)
			tx = fundsTx
		case p2p.STAKETX_RES:
			var stakeTx *protocol.StakeTx
			stakeTx = stakeTx.Decode(payload)
			tx = stakeTx
		}

		conn.Close()
	}

	return tx
}

func processFundsTx(state map[[32]byte]uint64, block *protocol.Block) {
	// Account collecting the Fees
	if _, exists := state[block.Beneficiary]; !exists {
		state[block.Beneficiary] = 0
	}
	for _, txHash := range block.FundsTxData {
		tx := reqTx(p2p.FUNDSTX_REQ, txHash)
		fundsTx := tx.(*protocol.FundsTx)
		source := fundsTx.From
		dest := fundsTx.To
		// Add accounts in the map if they don't exist
		if _, exists := state[source]; !exists {
			state[source] = InitialBalance
		}
		if _, exists := state[dest]; !exists {
			state[dest] = InitialBalance
		}
		state[source] = state[source] - fundsTx.Fee - fundsTx.Amount
		state[dest] += fundsTx.Amount
		state[block.Beneficiary] += fundsTx.Fee
	}
}

func TestBasicConsolidationTx(t *testing.T) {
	cleanAndPrepare()
	chain := createTestChain(t)

	// Create a snapshot of the current state
	state := make(map[[32]byte]uint64)

	// Process all the blocks in the chain
	for _, block := range chain {
		// Skip empty blocks
		if block == nil {
			continue
		}
		processFundsTx(state, block)
		// process other TX
	}

	// The consolidation tx creates a snapshot of the system till a certain
	// block of which we have to keep track of
	lastBlockHash := chain[len(chain)-1].Hash
	consTx, err := protocol.ConstrConsolidationTx(0, state, lastBlockHash)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(consTx)
}