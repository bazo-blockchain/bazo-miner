package storage

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/bazo-blockchain/bazo-miner/protocol"
)

//In-memory, k/v storage is tested with the test below
func TestReadWriteDeleteTx(t *testing.T) {

	rand := rand.New(rand.NewSource(time.Now().Unix()))

	var hashFundsSlice []*protocol.FundsTx
	var hashAccSlice []*protocol.ContractTx
	var hashConfigSlice []*protocol.ConfigTx
	var hashStakeSlice []*protocol.StakeTx

	testsize := 1000

	loopMax := testsize
	for i := 0; i < loopMax; i++ {
		tx, _ := protocol.ConstrFundsTx(0x01, rand.Uint64()%100000+1, rand.Uint64()%10+1, uint32(i), accA.Address, accB.Address, &PrivKeyA, nil)
		WriteOpenTx(tx)
		hashFundsSlice = append(hashFundsSlice, tx)
	}

	loopMax = testsize
	for i := 0; i < 1000; i++ {
		tx, _, _ := protocol.ConstrContractTx(0, rand.Uint64()%100+1, &RootPrivKey, nil, nil)
		WriteOpenTx(tx)
		hashAccSlice = append(hashAccSlice, tx)
	}

	//Restricted to 256, because the number of configTxs is stored in a uint8 in blocks
	loopMax = 256
	for cnt := 0; cnt < loopMax; cnt++ {
		tx, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), uint8(rand.Uint32()%5+1), rand.Uint64()%2342873423, rand.Uint64()%1000+1, uint8(cnt), &RootPrivKey)
		hashConfigSlice = append(hashConfigSlice, tx)
		WriteOpenTx(tx)
	}

	loopMax = testsize
	for cnt := 0; cnt < loopMax; cnt++ {
		isStaking := false
		if math.Mod(float64(cnt), 2.00) == 1 {
			isStaking = true
		}
		tx, _ := protocol.ConstrStakeTx(0, uint64(cnt), isStaking, accA.Address, &PrivKeyA, &CommitmentKeyA.PublicKey)
		hashStakeSlice = append(hashStakeSlice, tx)
		WriteOpenTx(tx)
	}

	for _, tx := range hashFundsSlice {
		if ReadOpenTx(tx.Hash()) == nil {
			t.Errorf("Error writing transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashAccSlice {
		if ReadOpenTx(tx.Hash()) == nil {
			t.Errorf("Error writing transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashConfigSlice {
		if ReadOpenTx(tx.Hash()) == nil {
			t.Errorf("Error writing transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashStakeSlice {
		if ReadOpenTx(tx.Hash()) == nil {
			t.Errorf("Error writing transaction hash: %x\n", tx)
		}
	}

	//Read all open txs, received in random order
	opentxs := ReadAllOpenTxs()
	//Comparing the total number of txs should be enough
	lenTotalTxs := len(hashStakeSlice) + len(hashConfigSlice) + len(hashFundsSlice) + len(hashAccSlice)
	if len(opentxs) != lenTotalTxs {
		errorMsg := fmt.Sprintf("ReadAllOpenTxs() returned an invalid list of transactions\n"+
			" (open: %d, total %d)\n", len(opentxs), lenTotalTxs)
		t.Error(errorMsg)
	}

	//Deleting open txs
	for _, tx := range hashFundsSlice {
		DeleteOpenTx(tx)
	}

	for _, tx := range hashAccSlice {
		DeleteOpenTx(tx)
	}

	for _, tx := range hashConfigSlice {
		DeleteOpenTx(tx)
	}

	for _, tx := range hashStakeSlice {
		DeleteOpenTx(tx)
	}

	//Make sure all txs are actually deleted
	for _, tx := range hashFundsSlice {
		if ReadOpenTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashAccSlice {
		if ReadOpenTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashConfigSlice {
		if ReadOpenTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashStakeSlice {
		if ReadOpenTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}

	//Same with k/v-based closed tx storage
	for _, tx := range hashAccSlice {
		WriteClosedTx(tx)
	}

	for _, tx := range hashFundsSlice {
		WriteClosedTx(tx)
	}

	for _, tx := range hashConfigSlice {
		WriteClosedTx(tx)
	}

	for _, tx := range hashStakeSlice {
		WriteClosedTx(tx)
	}

	for _, tx := range hashAccSlice {
		if ReadClosedTx(tx.Hash()) == nil {
			t.Errorf("Error writing to k/v storage: %x\n", tx)
		}
	}

	for _, tx := range hashFundsSlice {
		if ReadClosedTx(tx.Hash()) == nil {
			t.Errorf("Error writing to k/v storage: %x\n", tx)
		}
	}

	for _, tx := range hashConfigSlice {
		if ReadClosedTx(tx.Hash()) == nil {
			t.Errorf("Error writing to k/v storage: %x\n", tx)
		}
	}

	for _, tx := range hashStakeSlice {
		if ReadClosedTx(tx.Hash()) == nil {
			t.Errorf("Error writing to k/v storage: %x\n", tx)
		}
	}

	//Delete transactions from closed storage
	for _, tx := range hashAccSlice {
		DeleteClosedTx(tx)
	}

	for _, tx := range hashFundsSlice {
		DeleteClosedTx(tx)
	}

	for _, tx := range hashConfigSlice {
		DeleteClosedTx(tx)
	}

	for _, tx := range hashStakeSlice {
		DeleteClosedTx(tx)
	}

	//Make sure all txs are actually deleted
	for _, tx := range hashAccSlice {
		if ReadClosedTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashFundsSlice {
		if ReadClosedTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashConfigSlice {
		if ReadClosedTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}

	for _, tx := range hashStakeSlice {
		if ReadClosedTx(tx.Hash()) != nil {
			t.Errorf("Error deleting transaction hash: %x\n", tx)
		}
	}
}

//Disk-based k/v storage
func TestReadWriteDeleteBlock(t *testing.T) {

	//No panic
	DeleteOpenBlock([32]byte{'0'})

	b, b2, b3 := new(protocol.Block), new(protocol.Block), new(protocol.Block)
	b.Hash = [32]byte{'0'}
	b2.Hash = [32]byte{'1'}
	b3.Hash = [32]byte{'2'}
	WriteOpenBlock(b)
	WriteOpenBlock(b2)
	WriteOpenBlock(b3)

	if ReadOpenBlock(b.Hash) == nil || ReadOpenBlock(b2.Hash) == nil || ReadOpenBlock(b3.Hash) == nil {
		t.Error("Failed to write block to open block storage.\n")
	}

	newb1 := ReadOpenBlock(b.Hash)
	newb2 := ReadOpenBlock(b2.Hash)
	newb3 := ReadOpenBlock(b3.Hash)

	DeleteOpenBlock(newb1.Hash)
	DeleteOpenBlock(newb2.Hash)
	DeleteOpenBlock(newb3.Hash)

	WriteClosedBlock(newb1)
	WriteClosedBlock(newb2)
	WriteClosedBlock(newb3)

	if ReadOpenBlock(newb1.Hash) != nil ||
		ReadOpenBlock(newb2.Hash) != nil ||
		ReadOpenBlock(newb3.Hash) != nil ||
		ReadClosedBlock(b.Hash) == nil ||
		ReadClosedBlock(b2.Hash) == nil ||
		ReadClosedBlock(b3.Hash) == nil {
		t.Error("Failed to write block to closed block storage.\n")
	}

	DeleteClosedBlock(newb1.Hash)
	DeleteClosedBlock(newb2.Hash)
	DeleteClosedBlock(newb3.Hash)

	if ReadClosedBlock(b.Hash) != nil ||
		ReadClosedBlock(b2.Hash) != nil ||
		ReadClosedBlock(b3.Hash) != nil {
		t.Error("Failed to delete block from closed block storage.\n")
	}

	WriteLastClosedBlock(newb1)

	if ReadLastClosedBlock() == nil {
		t.Error("Failed to write block to last closed block storage.\n")
	}
	if !reflect.DeepEqual(newb1, ReadLastClosedBlock()) {
		t.Error("Failed to read last closed block from storage")
	}

	DeleteLastClosedBlock(newb1.Hash)

	if ReadLastClosedBlock() != nil {
		t.Error("Failed to delete last closed block from storage.\n")
	}

	WriteLastClosedBlock(newb1)

	if ReadLastClosedBlock() == nil {
		t.Error("Failed to write block to last closed block storage.\n")
	}

	DeleteAllLastClosedBlock()

	if ReadLastClosedBlock() != nil {
		t.Error("Failed to delete last closed block from storage.\n")
	}
}

func TestReadWriteDeleteEpochBlock(t *testing.T) {

	//No panic
	DeleteOpenEpochBlock([32]byte{'0'})

	//Initialize epoch blocks
	b, b2, b3 := new(protocol.EpochBlock), new(protocol.EpochBlock), new(protocol.EpochBlock)

	/*Set characteristics of epoch block 1*/
	var prevShardHashesEpochBlock1 [][32]byte
	var heightEpochBlock1 uint32
	//Assuming that the previous epoch had 2 running shards. Each hashX denotes the hash value of the last shard block
	hash11 := [32]byte{'0', '1'}
	hash12 := [32]byte{'0', '1'}
	prevShardHashesEpochBlock1 = append(prevShardHashesEpochBlock1, hash11)
	prevShardHashesEpochBlock1 = append(prevShardHashesEpochBlock1, hash12)
	heightEpochBlock1 = 100
	b.Hash = [32]byte{'0'}
	b.PrevShardHashes = prevShardHashesEpochBlock1
	b.Height = heightEpochBlock1

	/*Set characteristics of epoch block 2*/
	var prevShardHashesEpochBlock2 [][32]byte
	var heightEpochBlock2 uint32
	//Assuming that the previous epoch had 3 running shards. Each hashX denotes the hash value of the last shard block
	hash21 := [32]byte{'0', '1'}
	hash22 := [32]byte{'0', '1'}
	hash23 := [32]byte{'0', '1'}
	prevShardHashesEpochBlock2 = append(prevShardHashesEpochBlock1, hash21)
	prevShardHashesEpochBlock2 = append(prevShardHashesEpochBlock1, hash22)
	prevShardHashesEpochBlock2 = append(prevShardHashesEpochBlock1, hash23)
	heightEpochBlock2 = 100
	b2.Hash = [32]byte{'1'}
	b2.PrevShardHashes = prevShardHashesEpochBlock2
	b2.Height = heightEpochBlock2

	/*Set characteristics of epoch block 3*/
	var prevShardHashesEpochBlock3 [][32]byte
	var heightEpochBlock3 uint32
	//Assuming that the previous epoch had 3 running shards. Each hashX denotes the hash value of the last shard block
	hash31 := [32]byte{'0', '1'}
	hash32 := [32]byte{'0', '1'}
	hash33 := [32]byte{'0', '1'}
	hash34 := [32]byte{'0', '1'}
	prevShardHashesEpochBlock3 = append(prevShardHashesEpochBlock1, hash31)
	prevShardHashesEpochBlock3 = append(prevShardHashesEpochBlock1, hash32)
	prevShardHashesEpochBlock3 = append(prevShardHashesEpochBlock1, hash33)
	prevShardHashesEpochBlock3 = append(prevShardHashesEpochBlock1, hash34)
	heightEpochBlock3 = 100
	b3.Hash = [32]byte{'2'}
	b3.PrevShardHashes = prevShardHashesEpochBlock3
	b3.Height = heightEpochBlock3

	WriteOpenEpochBlock(b)
	WriteOpenEpochBlock(b2)
	WriteOpenEpochBlock(b3)

	if ReadOpenEpochBlock(b.Hash) == nil || ReadOpenEpochBlock(b2.Hash) == nil || ReadOpenEpochBlock(b3.Hash) == nil {
		t.Error("Failed to write epoch block to open block storage.\n")
	}

	newb1 := ReadOpenEpochBlock(b.Hash)
	newb2 := ReadOpenEpochBlock(b2.Hash)
	newb3 := ReadOpenEpochBlock(b3.Hash)

	DeleteOpenEpochBlock(newb1.Hash)
	DeleteOpenEpochBlock(newb2.Hash)
	DeleteOpenEpochBlock(newb3.Hash)

	WriteClosedEpochBlock(newb1)
	WriteClosedEpochBlock(newb2)
	WriteClosedEpochBlock(newb3)

	if ReadOpenEpochBlock(newb1.Hash) != nil ||
		ReadOpenEpochBlock(newb2.Hash) != nil ||
		ReadOpenEpochBlock(newb3.Hash) != nil ||
		ReadClosedEpochBlock(b.Hash) == nil ||
		ReadClosedEpochBlock(b2.Hash) == nil ||
		ReadClosedEpochBlock(b3.Hash) == nil {
		t.Error("Failed to write epoch block to closed block storage.\n")
	}

	DeleteClosedEpochBlock(newb1.Hash)
	DeleteClosedEpochBlock(newb2.Hash)
	DeleteClosedEpochBlock(newb3.Hash)

	if ReadClosedEpochBlock(b.Hash) != nil ||
		ReadClosedEpochBlock(b2.Hash) != nil ||
		ReadClosedEpochBlock(b3.Hash) != nil {
		t.Error("Failed to delete block from closed block storage.\n")
	}
}

func TestReadWriteDeleteState(t *testing.T) {

	newState := new(protocol.State)
	newState.ActualState = State

	WriteState(newState)

	_, err := ReadClosedState()

	if err != nil{
		t.Error("Failed to write state to closed storage.\n")
	}

	DeleteState()

	retrievedState, _ := ReadClosedState()

	if retrievedState != nil {
		t.Error("Failed to delete block from closed block storage.\n")
	}
}