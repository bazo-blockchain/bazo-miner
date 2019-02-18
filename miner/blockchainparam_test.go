package miner

import (
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"testing"
)

//Testing whether target calculation responds to block rollbacks
func TestTargetHistory(t *testing.T) {
	cleanAndPrepare()

	activeParameters.Diff_interval = 5
	activeParameters.Block_interval = 5

	//Build 5 blocks, this results in a targets update and a targetTime update
	//with timerange.first = 0 because of the genesis block
	var blocks []*protocol.Block
	var tmpBlock *protocol.Block
	tmpBlock = new(protocol.Block)
	for cnt := 0; cnt < 10; cnt++ {
		if(cnt == 0){
			tmpBlock = newBlock(lastBlock.HashBlock(), [crypto.COMM_PROOF_LENGTH]byte{}, 2)
		} else {
			tmpBlock = newBlock(tmpBlock.Hash, [crypto.COMM_KEY_LENGTH]byte{}, tmpBlock.Height+1)
		}
		finalizeBlock(tmpBlock)
		validate(tmpBlock, false)
		blocks = append(blocks, tmpBlock)
	}

	//Temporarily save the last target time to test after rollback
	tmpTimeRange := timerange{
		targetTimes[len(targetTimes)-1].first,
		targetTimes[len(targetTimes)-1].last,
	}

	//Make sure the arrays get expanded and contracted when they should
	var targetSize, targetTimesSize int

	targetSize = len(target)
	targetTimesSize = len(targetTimes)

	//This rollback causes the previous target and timerange to get active again
	rollback(blocks[len(blocks)-1])
	blocks = blocks[:len(blocks)-1]

	if targetSize == len(target) || targetTimesSize == len(targetTimes) {
		t.Error("Arrays for target change have not been updated.\n")
	}

	//The previous timerange needs the first value to be set and the the last value set to zero
	if currentTargetTime.last != 0 || currentTargetTime.first != tmpTimeRange.first {
		t.Error("Target time rollback failed.\n")
	}

	targetSize = len(target)
	targetTimesSize = len(targetTimes)

	tmpBlock = newBlock(blocks[len(blocks)-1].Hash, [crypto.COMM_PROOF_LENGTH]byte{}, blocks[len(blocks)-1].Height+1)
	finalizeBlock(tmpBlock)
	validate(tmpBlock, false)

	if targetSize == len(target) || targetTimesSize == len(targetTimes) {
		t.Error("Arrays for target change have not been updated.\n")
	}
}

//Tests whether system changes of relevant parameters influence the code
func TestTimestamps(t *testing.T) {
	cleanAndPrepare()

	//tweak parameters to test target update
	activeParameters.Diff_interval = 5
	activeParameters.Block_interval = 10

	prevHash := lastBlock.HashBlock()
	for cnt := 0; cnt < 0; cnt++ {
		//b := newBlock(prevHash, [crypto.COMM_PROOF_LENGTH]byte{}, 1)
		b := newBlock(prevHash, [crypto.COMM_PROOF_LENGTH]byte{}, 2)

		if cnt == 8 {
			tx, err := protocol.ConstrConfigTx(0, protocol.DIFF_INTERVAL_ID, 20, 2, 0, PrivKeyRoot)
			tx2, err2 := protocol.ConstrConfigTx(0, protocol.BLOCK_INTERVAL_ID, 60, 2, 0, PrivKeyRoot)
			if err != nil || err2 != nil {
				t.Errorf("Creating config txs failed: %v, %v\n", err, err2)
			}
			err = addTx(b, tx)
			err2 = addTx(b, tx2)
			if err != nil || err2 != nil {
				t.Errorf("Adding config txs to the block failed: %v, %v\n", err, err2)
			}
		}
		finalizeBlock(b)
		validate(b, false)
		prevHash = b.Hash

		//block is validated, check if configtx are now in the system
		if cnt == 8 {
			if activeParameters.Block_interval != 60 || activeParameters.Diff_interval != 20 || localBlockCount != 0 {
				t.Errorf("Block Interval: %v, Diff Interval: %v, LocalBlockCnt: %v\n",
					activeParameters.Block_interval,
					activeParameters.Diff_interval,
					localBlockCount,
				)
			}
		}
	}
}

//Tests whether the diff logic respects edge cases
func TestCalculateNewDifficulty(t *testing.T) {
	cleanAndPrepare()

	//set new system parameters
	target[len(target)-1] = 10
	activeParameters.Block_interval = 10
	activeParameters.Diff_interval = 10
	time := timerange{0, 100}

	if calculateNewDifficulty(&time) != 10 {
		t.Errorf("Difficulty should: %v, difficulty is: %v\n", 10, calculateNewDifficulty(&time))
	}

	//test for illegal values
	time = timerange{100, 99}
	if calculateNewDifficulty(&time) != 10 {
		t.Errorf("Difficult should: %v, difficulty is: %v\n", 10, calculateNewDifficulty(&time))
	}

	//should: 100, is: 900, target should be -3
	time = timerange{100, 1000}
	if calculateNewDifficulty(&time) != getDifficulty()-3 {
		t.Errorf("Difficulty should: %v, difficulty is: %v\n", 7, calculateNewDifficulty(&time))
	}

	//should: 100, is: 500, log2(0.2) = -2.3 -> target -= 2
	time = timerange{100, 600}
	if calculateNewDifficulty(&time) != getDifficulty()-2 {
		t.Errorf("Difficulty should: %v, difficulty is: %v\n", 8, calculateNewDifficulty(&time))
	}

	//should: 100, is: 1, log2(100) > 3 -> target_change = 3
	time = timerange{1000, 1001}
	if calculateNewDifficulty(&time) != getDifficulty()+3 {
		t.Errorf("Difficulty should: %v, difficulty is: %v\n", 13, calculateNewDifficulty(&time))
	}

	//should: 100, is: 50, log2(2) = 1
	time = timerange{100, 150}
	if calculateNewDifficulty(&time) != getDifficulty()+1 {
		t.Errorf("Difficulty should: %v, difficulty is: %v\n", 11, calculateNewDifficulty(&time))
	}
}
