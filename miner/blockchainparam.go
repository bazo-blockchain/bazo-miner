package miner

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"math"
	"time"
)

var (
	lastBlock         *protocol.Block
	blockBeingProcessed *protocol.Block
	blocksReceived 		[]*protocol.Block
	lastEpochBlock	  *protocol.EpochBlock
	firstEpochOver	  bool
	globalBlockCount  = int64(-1)
	localBlockCount   = int64(-1)
	target            []uint8    //Stores the history of target values
	currentTargetTime *timerange //Corresponds to the active timerange
	FirstEpochBlock	  *protocol.EpochBlock
)

//An instance of this datastructure is created whenever system parameters change.
//The blockhash is additionally recorded to know which blocks the parameter change belongs to.
//This is necessary, because the system records ALL config txs (even those who have no corresponding
//code to execute [e.g., when they're running an older version of the code]).
type Parameters struct {
	BlockHash               	[BLOCKHASH_SIZE]byte
	Fee_minimum             	uint64 //Paid minimum fee for sending a tx.
	Block_size              	uint64 //Block size in bytes.
	Diff_interval           	uint64
	Block_interval          	uint64
	Block_reward            	uint64 //Reward for delivering the correct PoS.
	Staking_minimum         	uint64 //Minimum amount a validator must own for staking.
	Waiting_minimum         	uint64 //Number of blocks that must a new validator must wait before it can start validating.
	Accepted_time_diff      	uint64 //Number of seconds that a block can be received in the future.
	Slashing_window_size    	uint64 //Number of blocks that a validator cannot vote on two competing chains.
	Slash_reward            	uint64 //Reward for providing the correct slashing proof.
	num_included_prev_proofs	int
	epoch_length				uint64
	validators_per_shard		uint64
}

func NewDefaultParameters() Parameters {
	newParameters := Parameters{
		[BLOCKHASH_SIZE]byte{},
		FEE_MINIMUM,
		BLOCK_SIZE,
		DIFF_INTERVAL,
		BLOCK_INTERVAL,
		BLOCK_REWARD,
		STAKING_MINIMUM,
		WAITING_MINIMUM,
		ACCEPTED_TIME_DIFF,
		SLASHING_WINDOW_SIZE,
		SLASH_REWARD,
		NUM_INCL_PREV_PROOFS,
		EPOCH_LENGTH,
		VALIDATORS_PER_SHARD,
	}

	return newParameters
}

//Captures first and last timestamp of the intended blocks of the range.
type timerange struct {
	first int64
	last  int64
}

//We need to store the history or timeranges to revert in case of rollbacks.
var targetTimes []timerange

func collectStatistics(b *protocol.Block) {
	logger.Printf("--- BLOCK STATISTICS FOR BLOCK (%x) AT HEIGHT: %d ---\n", b.Hash[0:8],b.Height)
	FileConnectionsLog.WriteString(fmt.Sprintf("--- BLOCK STATISTICS FOR BLOCK (%x) AT HEIGHT: %d ---\n", b.Hash[0:8],b.Height))

	blockEndTime = time.Now().Unix()
	var blockDuration = blockEndTime - blockStartTime

	globalBlockCount++
	localBlockCount++
	totalTransactionsInBlock := b.NrFundsTx + uint16(b.NrConfigTx) + b.NrContractTx + b.NrStakeTx
	validatedTXCount = validatedTXCount + int(totalTransactionsInBlock)
	validatedBlockCount = validatedBlockCount + 1

	logger.Printf("Number of transactions in block (%x): %d\n", b.Hash[0:8],totalTransactionsInBlock)
	FileConnectionsLog.WriteString(fmt.Sprintf("Number of transactions in block: %d\n", totalTransactionsInBlock))

	logger.Printf("Validated Transaction count: %d\n", validatedTXCount)
	FileConnectionsLog.WriteString(fmt.Sprintf("Validated Transaction count: %d\n", validatedTXCount))

	logger.Printf("Validated Block count: %d\n", validatedBlockCount)
	FileConnectionsLog.WriteString(fmt.Sprintf("Validated Block count: %d\n", validatedBlockCount))

	logger.Printf("Block done at time: %d\n", time.Now().Unix())
	FileConnectionsLog.WriteString(fmt.Sprintf("Block done at time: %d\n", time.Now().Unix()))

	logger.Printf("Block duration: %d seconds\n",blockDuration)
	FileConnectionsLog.WriteString(fmt.Sprintf("Block duration: %d seconds\n",blockDuration))

	var blockTPS = float64(totalTransactionsInBlock) / float64(blockEndTime-blockStartTime)

	logger.Printf("Block TPS: %v TX/Sec\n", blockTPS)
	FileConnectionsLog.WriteString(fmt.Sprintf("Block TPS: %v TX/Sec\n", blockTPS))

	logger.Printf("MemPool Size: %d\n", storage.GetMemPoolSize())
	FileConnectionsLog.WriteString(fmt.Sprintf("MemPool Size: %d\n", storage.GetMemPoolSize()))

	if localBlockCount == int64(activeParameters.Diff_interval) {
		currentTargetTime.last = b.Timestamp
		//The genesis block has timestamp = 0. This simplifies certain things: Every miner can start with an already
		//existing genesis block (because all fields are set to 0). The "find common ancestor" algorithm can then
		//use the genesis block as a common ancestor for new miners who have not synchronized with the chain yet.
		if currentTargetTime.first == 0 {
			target = append(target, target[len(target)-1])
		} else {
			target = append(target, calculateNewDifficulty(currentTargetTime))
		}

		targetTimes = append(targetTimes, *currentTargetTime)

		logger.Printf("Target changed, new target: %v\n", target[len(target)-1])
		FileConnectionsLog.WriteString(fmt.Sprintf("Target changed, new target: %v\n", target[len(target)-1]))
		localBlockCount = 0
		currentTargetTime = new(timerange)
		currentTargetTime.first = b.Timestamp
	}

	lastBlock = b
}

func collectStatisticsRollback(b *protocol.Block) {
	globalBlockCount--

	//Never rollback the genesis blocks.
	if localBlockCount == 0 && globalBlockCount != 0 {
		localBlockCount = int64(activeParameters.Diff_interval) - 1
		//Target rollback
		target = target[:len(target)-1]
		currentTargetTime.first = targetTimes[len(targetTimes)-1].first
		targetTimes = targetTimes[:len(targetTimes)-1]
	} else {
		localBlockCount--
	}

	lastBlock = storage.ReadClosedBlock(b.PrevHash)
}

func calculateNewDifficulty(t *timerange) uint8 {
	//Time difference between the first and last block in the measured range.
	diff_now := t.last - t.first

	//This is how long it should have taken.
	diff_wanted := activeParameters.Block_interval * (activeParameters.Diff_interval)

	diff_ratio := float64(diff_wanted) / float64(diff_now)

	//If the last is earlier time than first, we get a negative number, can't take the log from that.
	//This precipitates that reasonable parameter should be chosen for block-/diff interval
	//such that this case does not happen. In case it still does, we give the current difficulty back.
	if diff_ratio < 0 {
		return getDifficulty()
	}

	//Take the log2 from the diff_ratio, because adding a zero makes it twice as hard, adding two zeros four times as
	//hard etc.
	target_change := math.Log2(diff_ratio)

	//The +-0.5 is basically the "round" function.
	if target_change > 0 {
		target_change += 0.5
	} else if target_change < 0 {
		target_change -= 0.5
	}

	//Sanity check! Make it at most 3 times as hard or easy, Bitcoin has a similar check.
	if target_change > 3 {
		target_change = 3
	} else if target_change < -3 {
		target_change = -3
	}

	//Rounding down (for positive values) and runding up (for negative values).
	target_change_rounded := uint8(target_change)

	//Return the new target based on the calculation and the current target.
	return target_change_rounded + target[len(target)-1]
}

func getDifficulty() uint8 {
	return target[len(target)-1]
}

func (param Parameters) String() string {
	return fmt.Sprintf(
		"\n"+
			"Block Hash: %x\n"+
			"Block size: %v\n"+
			"Difficulty interval: %v\n"+
			"Fee minimum: %v\n"+
			"Block interval: %v\n"+
			"Block reward: %v\n"+
			"Staking minimum: %v\n"+
			"Waiting minimum: %v\n"+
			"Acceptanced time difference: %v\n"+
			"Slashing window size: %v\n"+
			"Slash reward: %v\n"+
			"Num of previous proofs included in PoS: %v\n"+
			"Epoch Length: %v\n"+
			"Validators per Shard: %v\n",
		param.BlockHash[0:8],
		param.Block_size,
		param.Diff_interval,
		param.Fee_minimum,
		param.Block_interval,
		param.Block_reward,
		param.Staking_minimum,
		param.Waiting_minimum,
		param.Accepted_time_diff,
		param.Slashing_window_size,
		param.Slash_reward,
		param.num_included_prev_proofs,
		param.epoch_length,
		param.validators_per_shard,
	)
}
