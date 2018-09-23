package miner

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

//Tests block adding, verification, serialization and deserialization
//This test goes further than protocol/block_test.go because it tests the integrity of the payloads as well
//while protocol/block_test.go only tests serialization/deserialization and size calculation
func TestBlock(t *testing.T) {
	cleanAndPrepare()

	b := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	hashFundsSlice, hashAccSlice, hashConfigSlice, hashStakeSlice := createBlockWithTxs(b)
	err := finalizeBlock(b)
	if err != nil {
		t.Errorf("Block finalization failed (%v)\n", err)
		return
	}

	encodedBlock := b.Encode()

	var decodedBlock *protocol.Block

	decodedBlock = decodedBlock.Decode(encodedBlock)

	err = validate(decodedBlock, false)

	b.StateCopy = nil
	decodedBlock.StateCopy = nil

	if err != nil {
		t.Errorf("Block validation failed (%v)\n", err)
	}
	if !reflect.DeepEqual(hashFundsSlice, decodedBlock.FundsTxData) {
		t.Error("FundsTx data is not properly serialized!")
	}
	if !reflect.DeepEqual(hashAccSlice, decodedBlock.AccTxData) {
		t.Error("AccTx data is not properly serialized!")
	}
	if !reflect.DeepEqual(hashConfigSlice, decodedBlock.ConfigTxData) {
		t.Error("ConfigTx data is not properly serialized!")
	}
	if !reflect.DeepEqual(hashStakeSlice, decodedBlock.StakeTxData) {
		t.Error("StakeTx data is not properly serialized!")
	}
	if !reflect.DeepEqual(b, decodedBlock) {
		t.Error("Either serialization or deserialization failed, blocks are not equal!")
	}
}

//Duplicate Txs are not allowed
func TestBlockTxDuplicates(t *testing.T) {

	cleanAndPrepare()
	b := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	createBlockWithTxs(b)

	if err := finalizeBlock(b); err != nil {
		t.Errorf("Block finalization failed. (%v)\n", err)
	}

	//This is a normal block validation, should pass
	if err := validate(b, false); err != nil {
		t.Errorf("Block validation failed. (%v)\n", err)
	}
	t.Log(lastBlock)

	//Rollback the block and add a duplicate
	err := rollback(b)
	if err != nil {
		t.Log(err)
	}
	t.Log(lastBlock)

	if len(b.ConfigTxData) > 0 {
		b.ConfigTxData = append(b.ConfigTxData, b.ConfigTxData[0])
	}

	if err := finalizeBlock(b); err != nil {
		t.Errorf("Block finalization failed. (%v)\n", err)
	}

	if err := validate(b, false); err == nil {
		t.Errorf("Duplicate Tx not detected.\n")
	}
	t.Log(lastBlock)

}

//Blocks that link to the previous block and have valid txs should pass
func TestMultipleBlocks(t *testing.T) {
	cleanAndPrepare()

	b := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	createBlockWithTxs(b)
	finalizeBlock(b)
	if err := validate(b, false); err != nil {
		t.Errorf("Block validation for (%v) failed: %v\n", b, err)
	}

	b2 := newBlock(b.Hash, [32]byte{}, [32]byte{}, 2)
	createBlockWithTxs(b2)
	finalizeBlock(b2)
	if err := validate(b2, false); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}

	b3 := newBlock(b2.Hash, [32]byte{}, [32]byte{}, 3)
	createBlockWithTxs(b3)
	finalizeBlock(b3)
	if err := validate(b3, false); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}

	b4 := newBlock(b3.Hash, [32]byte{}, [32]byte{}, 4)
	createBlockWithTxs(b4)
	finalizeBlock(b4)
	if err := validate(b4, false); err != nil {
		t.Errorf("Block validation failed: %v\n", err)
	}
}

//Test the blocktimestamp check
func TestTimestampCheck(t *testing.T) {
	cleanAndPrepare()

	timePast := time.Now().Unix() - 4000
	timeFuture := time.Now().Unix() + 4000
	timeNow := time.Now().Unix() + 50

	if err := timestampCheck(timePast); err == nil {
		t.Error("Dynamic time check failed\n")
	}

	if err := timestampCheck(timeFuture); err == nil {
		t.Error("Dynamic time check failed\n")
	}

	if err := timestampCheck(timeNow); err != nil {
		t.Errorf("Valid time got rejected: %v\n", err)
	}
}

//Helper function used by lots of test to fill the block with some random data
func createBlockWithTxs(b *protocol.Block) ([][32]byte, [][32]byte, [][32]byte, [][32]byte) {
	var testSize uint32
	testSize = 100

	var hashFundsSlice [][32]byte
	var hashAccSlice [][32]byte
	var hashConfigSlice [][32]byte
	var hashStakeSlice [][32]byte

	//in order to create valid funds transactions we need to know the tx count of acc A
	randVar := rand.New(rand.NewSource(time.Now().Unix()))
	loopMax := int(randVar.Uint32()%testSize) + 1
	loopMax += int(accA.TxCnt)
	for cnt := int(accA.TxCnt); cnt < loopMax; cnt++ {
		accAHash := protocol.SerializeHashContent(accA.Address)
		accBHash := protocol.SerializeHashContent(accB.Address)
		tx, _ := protocol.ConstrFundsTx(0x01, randVar.Uint64()%100+1, randVar.Uint64()%100+1, uint32(cnt), accAHash, accBHash, &PrivKeyAccA, &PrivKeyMultiSig)
		if err := addTx(b, tx); err == nil {
			//Might  be that we generated a block that was already generated before
			if storage.ReadOpenTx(tx.Hash()) != nil || storage.ReadClosedTx(tx.Hash()) != nil {
				continue
			}
			hashFundsSlice = append(hashFundsSlice, tx.Hash())
			storage.WriteOpenTx(tx)
		} else {
			fmt.Print(err)
		}
	}

	nullAddress := [64]byte{}
	loopMax = int(randVar.Uint32()%testSize) + 1
	for cnt := 0; cnt < loopMax; cnt++ {
		tx, _, _ := protocol.ConstrAccTx(0, randVar.Uint64()%100+1, nullAddress, &PrivKeyRoot)
		if err := addTx(b, tx); err == nil {
			if storage.ReadOpenTx(tx.Hash()) != nil || storage.ReadClosedTx(tx.Hash()) != nil {
				continue
			}
			hashAccSlice = append(hashAccSlice, tx.Hash())
			storage.WriteOpenTx(tx)
		} else {
			fmt.Print(err)
		}
	}

	//NrConfigTx is saved in a uint8, so testsize shouldn't be larger than 255
	loopMax = int(randVar.Uint32()%testSize) + 1
	for cnt := 0; cnt < loopMax; cnt++ {
		tx, err := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), uint8(randVar.Uint32()%10+1), randVar.Uint64()%2342873423, randVar.Uint64()%1000+1, uint8(cnt), &PrivKeyRoot)
		if err != nil {
			fmt.Print(err)
		}
		if storage.ReadOpenTx(tx.Hash()) != nil || storage.ReadClosedTx(tx.Hash()) != nil {
			continue
		}

		//don't mess with the minimum fee, block size and staking minimum
		if tx.Id == 3 || tx.Id == 1 || tx.Id == 6 {
			continue
		}
		if err := addTx(b, tx); err == nil {

			hashConfigSlice = append(hashConfigSlice, tx.Hash())
			storage.WriteOpenTx(tx)
		} else {
			fmt.Print(err)
		}
	}

	return hashFundsSlice, hashAccSlice, hashConfigSlice, hashStakeSlice
}

//TODO
//func TestReadLastClosedBlock(t *testing.T) {
//	cleanAndPrepare()
//
//	lastClosedBlock := storage.ReadLastClosedBlock()
//
//	if !reflect.DeepEqual(lastClosedBlock, genesisBlock) {
//		t.Errorf("Genesis Block is not read as a closed block:\n%v\n%v", lastClosedBlock, genesisBlock)
//		return
//	}
//
//	var lastClosedBlocksAfterGenesis []*protocol.Block
//	lastClosedBlocksAfterGenesis = append(lastClosedBlocksAfterGenesis, genesisBlock)
//
//	lastClosedBlocks := storage.ReadAllClosedBlocks()
//	if !reflect.DeepEqual(lastClosedBlocks, lastClosedBlocksAfterGenesis) {
//		t.Errorf("Closed blocks are not equal after genesis block:\n%v\n%v", lastClosedBlocks, lastClosedBlocksAfterGenesis)
//	}
//}
