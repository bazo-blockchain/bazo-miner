package miner

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

var (
	logger                  *log.Logger
	FileLogger                  *log.Logger
	blockValidation         = &sync.Mutex{}
	parameterSlice          []Parameters
	activeParameters        *Parameters
	uptodate                bool
	prevBlockIsEpochBlock   bool
	FirstStartAfterEpoch	bool
	slashingDict            = make(map[[64]byte]SlashingProof)
	validatorAccAddress     [64]byte
	commPrivKey             *rsa.PrivateKey
	NumberOfShards          int
	LastShardHashes         [][32]byte // This slice stores the hashes of the last blocks from the other shards, needed to create the next epoch block
	ValidatorShardMap       *protocol.ValShardMapping // This map keeps track of the validator assignment to the shards; int: shard ID; [64]byte: validator address
	FileConnections   	       *os.File
	FileConnectionsLog         *os.File
	TransactionPayloadOut 	*protocol.TransactionPayload
	validatedTXCount		int
	validatedBlockCount		int
	blockStartTime			int64
	syncStartTime			int64
	blockEndTime			int64
	totalSyncTime			int64
)

func InitFirstStart(wallet *ecdsa.PublicKey, commitment *rsa.PrivateKey) error {
	var err error
	FileConnections, err = os.OpenFile(fmt.Sprintf("hash-prevhash-%v.txt",strings.Split(p2p.Ipport, ":")[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	FileConnectionsLog, err = os.OpenFile(fmt.Sprintf("hlog-for-%v.txt",strings.Split(p2p.Ipport, ":")[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	logger = storage.InitLogger()
	FileLogger = storage.InitFileLogger()
	FileLogger.SetOutput(FileConnectionsLog)

	rootAddress := crypto.GetAddressFromPubKey(wallet)

	var rootCommitment [crypto.COMM_KEY_LENGTH]byte
	copy(rootCommitment[:], commitment.N.Bytes())

	genesis := protocol.NewGenesis(rootAddress, rootCommitment)
	storage.WriteGenesis(&genesis)

	//logger.Printf("Written Genesis Block: %v\n", genesis.String())
	//FileConnectionsLog.WriteString(fmt.Sprintf("Written Genesis Block: %v\n", genesis.String()))

	/*Write First Epoch block chained to the genesis block*/
	initialEpochBlock := protocol.NewEpochBlock([][32]byte{genesis.Hash()}, 0)
	initialEpochBlock.Hash = initialEpochBlock.HashEpochBlock()
	FirstEpochBlock = initialEpochBlock
	initialEpochBlock.State = storage.State
	storage.WriteFirstEpochBlock(initialEpochBlock)

	storage.WriteClosedEpochBlock(initialEpochBlock)

	storage.DeleteAllLastClosedEpochBlock()
	storage.WriteLastClosedEpochBlock(initialEpochBlock)

	FileLogger.Printf("Last Epoch block hash: (%x)\n",storage.ReadLastClosedEpochBlock().Hash)

	firstValMapping := protocol.NewMapping()
	initialEpochBlock.ValMapping = firstValMapping
	hashGenesis := [32]byte{}
	FileConnections.WriteString(fmt.Sprintf(`"GENESIS \n Hash : %x" -> "EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x"`+"\n", hashGenesis[0:8],initialEpochBlock.Hash[0:8],initialEpochBlock.Height,initialEpochBlock.MerklePatriciaRoot[0:8]))
	FileConnections.WriteString(fmt.Sprintf(`"GENESIS \n Hash : %x"`+`[color = green, shape = hexagon]`+"\n",hashGenesis[0:8]))
	FileConnections.WriteString(fmt.Sprintf(`"EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x"`+`[color = red, shape = box]`+"\n",initialEpochBlock.Hash[0:8],initialEpochBlock.Height,initialEpochBlock.MerklePatriciaRoot[0:8]))

	return Init(wallet, commitment)
}

//Miner entry point
func Init(wallet *ecdsa.PublicKey, commitment *rsa.PrivateKey) error {
	//this bool indicates whether the first epoch is over. Only in the first epoch, the bootstrapping node is assigning the
	//validators to the shards and broadcasts this assignment to the other miners
	firstEpochOver = false

	FileConnections, _ = os.OpenFile(fmt.Sprintf("hash-prevhash-%v.txt",strings.Split(p2p.Ipport, ":")[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	FileConnectionsLog, _ = os.OpenFile(fmt.Sprintf("hlog-for-%v.txt",strings.Split(p2p.Ipport, ":")[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	validatorAccAddress = crypto.GetAddressFromPubKey(wallet)
	commPrivKey = commitment

	//Set up logger.
	logger = storage.InitLogger()
	FileLogger = storage.InitFileLogger()
	FileLogger.SetOutput(FileConnectionsLog)


	parameterSlice = append(parameterSlice, NewDefaultParameters())
	activeParameters = &parameterSlice[0]

	currentTargetTime = new(timerange)
	target = append(target, 15)

	var initialBlock *protocol.Block
	var err error

	//Start to listen to network inputs (txs and blocks).
	go incomingData()
	go incomingEpochData()
	go incomingStateData()

	//Since new validators only join after the currently running epoch ends, they do no need to download the whole shardchain history,
	//but can continue with their work after the next epoch block and directly set their state to the global state of the first received epoch block
	if(p2p.IsBootstrap()){
		initialBlock, err = initState() //From here on, every validator should have the same state representation
		if err != nil {
			return err
		}
		FileConnections.WriteString(fmt.Sprintf(`"Hash : %x \n Height : %d" -> "Hash : %x \n Height : %d"`+"\n", initialBlock.PrevHash[0:8],initialBlock.Height-1,initialBlock.Hash[0:8],initialBlock.Height))
		lastBlock = initialBlock
	} else {
		for{
			//Wait until I receive the last epoch block as well as the validator assignment
			// The global variables 'lastEpochBlock' and 'ValidatorShardMap' are being set when they are received by the network
			if(lastEpochBlock != nil && ValidatorShardMap != nil){
				if(lastEpochBlock.Height > 0){
					storage.State = lastEpochBlock.State
					NumberOfShards = lastEpochBlock.NofShards
					storage.ThisShardID = ValidatorShardMap.ValMapping[validatorAccAddress] //Save my ShardID
					FirstStartAfterEpoch = true

					lastBlock = dummyLastBlock
					epochMining(lastEpochBlock.Hash,lastEpochBlock.Height) //start mining based on the received Epoch Block
				}
			}
		}
	}

	logger.Printf("Active config params:%v\n", activeParameters)
	FileLogger.Printf("Active config params:%v\n", activeParameters)

	/*Sharding Utilities*/
	NumberOfShards = DetNumberOfShards()

	/*First validator assignment is done by the bootstrapping node, the others will be done based on POS at the end of each epoch*/
	if (p2p.IsBootstrap()) {
		var validatorShardMapping = protocol.NewMapping()
		validatorShardMapping.ValMapping = AssignValidatorsToShards()
		validatorShardMapping.EpochHeight = int(lastEpochBlock.Height)
		ValidatorShardMap = validatorShardMapping
		storage.WriteValidatorMapping(ValidatorShardMap)
		logger.Printf("Validator Shard Mapping:\n")
		logger.Printf(validatorShardMapping.String())
		FileLogger.Print("Validator Shard Mapping:\n")
		FileLogger.Print(validatorShardMapping.String()+"\n")
	}

	storage.ThisShardID = ValidatorShardMap.ValMapping[validatorAccAddress]

	epochMining(lastBlock.Hash, lastBlock.Height)

	return nil
}

func epochMining(hashPrevBlock [32]byte, heightPrevBlock uint32) {

	var epochBlock *protocol.EpochBlock

	/*constantly watch height of the lastblock variable for insertion of next epoch block*/
	for {

		if(FirstStartAfterEpoch == true){ //Indicates that a validator newly joined Bazo after the current epoch, thus his 'lastBlock' variable is nil
			mining(hashPrevBlock,heightPrevBlock)
		}

		/*This is how it is done with state transition*/
		logger.Printf("Before checking my state stash for lastblock height: %d\n",lastBlock.Height)
		FileLogger.Printf("Before checking my state stash for lastblock height: %d\n",lastBlock.Height)
		syncStartTime = time.Now().Unix()

		shardIDs := makeRange(1,NumberOfShards) // get all shard ids from the current epoch
		FileLogger.Printf("Number of shards: %d\n",NumberOfShards)
		shardIDStateBoolMap := make(map[int]bool) // check whether state transition of shard X has been processed
		for k, _ := range shardIDStateBoolMap {
			shardIDStateBoolMap[k] = false
		}

		for{

			if(NumberOfShards == 1){
				break
			}

			//time.Sleep(3 * time.Second)

			stateStashForHeight := protocol.ReturnStateTransitionForHeight(storage.ReceivedStateStash,lastBlock.Height)

			if(len(stateStashForHeight) != 0){
				//iterate through state transitions and apply them to local state, keep track of processed shards
				for _,st := range stateStashForHeight{
					if(shardIDStateBoolMap[st.ShardID] == false){
						storage.State = storage.ApplyRelativeState(storage.State,st.RelativeStateChange)
						//Delete transactions from mempool, which were validated by the other shards
						DeleteTransactionFromMempool(st.ContractTxData,st.FundsTxData,st.ConfigTxData,st.StakeTxData)

						shardIDStateBoolMap[st.ShardID] = true

						FileLogger.Printf("Processed state transition of shard: %d\n",st.ShardID)
					}
				}

				if (len(stateStashForHeight) == NumberOfShards-1){
					break
				}
			}

			//Iterate over shard IDs to check which ones are still missing
			for _,id := range shardIDs{
				if(id != storage.ThisShardID && shardIDStateBoolMap[id] == false){
					var stateTransition *protocol.StateTransition

					FileLogger.Printf("requesting state transition for lastblock height: %d\n",lastBlock.Height)

					p2p.StateTransitionReqShard(id,int(lastBlock.Height))
					//Blocking wait
					select {
					case encodedStateTransition := <-p2p.StateTransitionShardReqChan:
						stateTransition = stateTransition.DecodeTransition(encodedStateTransition)

						storage.State = storage.ApplyRelativeState(storage.State,stateTransition.RelativeStateChange)
						shardIDStateBoolMap[stateTransition.ShardID] = true

						FileLogger.Printf("Writing state back to stash Shard ID: %v  VS my shard ID: %v - Height: %d\n",stateTransition.ShardID,storage.ThisShardID,stateTransition.Height)
						storage.ReceivedStateStash.Set(stateTransition.HashTransition(),stateTransition)

						//Limit waiting time to BLOCKFETCH_TIMEOUT seconds before aborting.
					case <-time.After(5 * time.Second):
						FileLogger.Printf("have been waiting for 5 seconds for lastblock height: %d\n",lastBlock.Height)
						continue
					}
				}
			}
		}
		logger.Printf("After checking my state stash for lastblock height: %d\n",lastBlock.Height)
		FileLogger.Printf("After checking my state stash for lastblock height: %d\n",lastBlock.Height)

		var syncEndTime = time.Now().Unix()
		var syncDuration = syncEndTime - syncStartTime
		totalSyncTime += syncDuration

		FileLogger.Printf("Synchronisation duration for lastblock height: %d - %d seconds\n",lastBlock.Height,syncDuration)
		FileLogger.Printf("Total Synchronisation duration for lastblock height: %d - %d seconds\n",lastBlock.Height,totalSyncTime)

		prevBlockIsEpochBlock = false // this boolean indicates whether the previous block is an epoch block

		//TODO: @Kürsat: lol, remove if clause
		if (true) {
			if (lastBlock.Height == uint32(lastEpochBlock.Height) + uint32(activeParameters.epoch_length)) {
				//The variable 'lastblock' is one before the next epoch block, thus with the lastblock, crete next epoch block

				epochBlock = protocol.NewEpochBlock([][32]byte{lastBlock.Hash}, lastBlock.Height+1)
				FileLogger.Printf("epochblock beingprocessed height: %d\n",epochBlock.Height)

				//if(NumberOfShards != 1 || storage.ReceivedBlockStashFromOtherShards != nil){
				if(NumberOfShards != 1){
					/*Extract the hashes of the last block of the other shards, needed to create the epoch block*/
					LastShardHashes = protocol.ReturnShardHashesForHeight(storage.ReceivedStateStash,lastBlock.Height)
					epochBlock.PrevShardHashes = append(epochBlock.PrevShardHashes,LastShardHashes...)
				}

				FileLogger.Printf("Before finalizeEpochBlock() ---- Height: %d\n",epochBlock.Height)
				//Include validator-shard mapping in the epoch block
				err := finalizeEpochBlock(epochBlock) //in case another epoch block was mined in the meantime, abort PoS here
				FileLogger.Printf("After finalizeEpochBlock() ---- Height: %d\n",epochBlock.Height)

				if err != nil {
					logger.Printf("%v\n", err)
					FileLogger.Printf("%v\n", err)
				} else {
					logger.Printf("EPOCH BLOCK mined (%x)\n", epochBlock.Hash[0:8])
					FileLogger.Printf("EPOCH BLOCK mined (%x)\n", epochBlock.Hash[0:8])
				}

				if err == nil {
					/*epochBlock.Hash = epochBlock.HashEpochBlock()
					epochBlock.State = storage.State*/

					//broadcast the generated map to the other validators
					/*broadcastValidatorShardMapping(ValidatorShardMap)*/
					//logger.Printf("Broadcast epoch block (%x)\n", epochBlock.Hash[0:8])
					FileLogger.Printf("Broadcast epoch block (%x)\n", epochBlock.Hash[0:8])
					broadcastEpochBlock(epochBlock)
					storage.WriteClosedEpochBlock(epochBlock)
					storage.DeleteAllLastClosedEpochBlock()
					storage.WriteLastClosedEpochBlock(epochBlock)
					lastEpochBlock = epochBlock

					//storage.WriteValidatorMapping(ValidatorShardMap)
					logger.Printf("Created Validator Shard Mapping :\n")
					logger.Printf(ValidatorShardMap.String())
					logger.Printf("Inserting EPOCH BLOCK: %v\n", epochBlock.String())
					FileLogger.Printf("Created Validator Shard Mapping :\n")
					FileLogger.Printf(ValidatorShardMap.String()+"\n")
					FileLogger.Printf("Inserting EPOCH BLOCK: %v\n", epochBlock.String())

					//broadcastEpochBlock(epochBlock)

					for _, prevHash := range epochBlock.PrevShardHashes {
						//FileConnections.WriteString(fmt.Sprintf("'%x' -> 'EPOCH BLOCK: %x'\n", prevHash[0:15], epochBlock.Hash[0:15]))
						FileConnections.WriteString(fmt.Sprintf(`"Hash : %x \n Height : %d" -> "EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x"`+"\n", prevHash[0:8],epochBlock.Height-1,epochBlock.Hash[0:8],epochBlock.Height,epochBlock.MerklePatriciaRoot[0:8]))
						FileConnections.WriteString(fmt.Sprintf(`"EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x"`+`[color = red, shape = box]`+"\n",epochBlock.Hash[0:8],epochBlock.Height,epochBlock.MerklePatriciaRoot[0:8]))
					}
				}

				prevBlockIsEpochBlock = true

				firstEpochOver = true

				mining(lastEpochBlock.Hash, lastEpochBlock.Height)
			} else if(lastEpochBlock.Height == lastBlock.Height+1){
				prevBlockIsEpochBlock = true
				mining(lastEpochBlock.Hash, lastEpochBlock.Height) //lastblock was received before we started creation of next epoch block
			} else {
				mining(lastBlock.Hash, lastBlock.Height)
			}
		}
	}
}

//Mining is a constant process, trying to come up with a successful PoW.
func mining(hashPrevBlock [32]byte, heightPrevBlock uint32) {

	blockStartTime = time.Now().Unix()
	currentBlock := newBlock(hashPrevBlock, [crypto.COMM_PROOF_LENGTH]byte{}, heightPrevBlock+1)

	/*Set shard identifier in block*/
	currentBlock.ShardId = storage.ThisShardID
	blockBeingProcessed = currentBlock

	logger.Printf("blockBeingProcessed Height: %v - MyShardID: %d\n",currentBlock.Height,storage.ThisShardID)
	FileLogger.Printf("blockBeingProcessed Height: %v - MyShardID: %d\n",currentBlock.Height,storage.ThisShardID)

	//This is the same mutex that is claimed at the beginning of a block validation. The reason we do this is
	//that before start mining a new block we empty the mempool which contains tx data that is likely to be
	//validated with block validation, so we wait in order to not work on tx data that is already validated
	//when we finish the block.
	blockValidation.Lock()
	FileLogger.Printf("Before preparing Block Height: %v\n",currentBlock.Height)
	prepareBlock(currentBlock) // In this step, filter tx from mem pool to check if they belong to my shard
	FileLogger.Printf("After preparing Block Height: %v\n",currentBlock.Height)
	blockValidation.Unlock()

	_, err := storage.ReadAccount(validatorAccAddress)
	if err != nil {
		logger.Printf("%v\n", err)
		FileLogger.Printf("%v\n", err)
		return
	}

	FileLogger.Printf("---- Before finalizeBlock() ---- Height: %d\n",currentBlock.Height)
	err = finalizeBlock(currentBlock) //in case another block was mined in the meantime, abort PoS here
	FileLogger.Printf("---- After finalizeBlock() ---- Height: %d\n",currentBlock.Height)

	if err != nil {
		logger.Printf("%v\n", err)
		FileLogger.Printf("%v\n", err)
	} else {
		logger.Printf("Block mined (%x)\n", currentBlock.Hash[0:8])
		FileLogger.Printf("Block mined (%x)\n", currentBlock.Hash[0:8])
	}

	if err == nil {
		err := validate(currentBlock, false) //here, block is written to closed storage and globalblockcount increased
		if err == nil{
			stateTransition := protocol.NewStateTransition(storage.RelativeState,int(currentBlock.Height),storage.ThisShardID,currentBlock.Hash,
				currentBlock.ContractTxData,currentBlock.FundsTxData,currentBlock.ConfigTxData,currentBlock.StakeTxData)

			FileLogger.Printf("Broadcast state transition for height %d\n", currentBlock.Height)
			broadcastStateTransition(stateTransition)
			storage.WriteToOwnStateTransitionkStash(stateTransition)
			//storage.OwnStateTransitionStash = append(storage.OwnStateTransitionStash,stateTransition)

			FileLogger.Printf("Broadcast block for height %d\n", currentBlock.Height)
			broadcastBlock(currentBlock)

			if (prevBlockIsEpochBlock == true || FirstStartAfterEpoch==true) {
				FileConnections.WriteString(fmt.Sprintf(`"EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x" -> "Hash : %x \n Height : %d"`+"\n", currentBlock.PrevHash[0:8],lastEpochBlock.Height,lastEpochBlock.MerklePatriciaRoot[0:8],currentBlock.Hash[0:8],currentBlock.Height))
				FileConnections.WriteString(fmt.Sprintf(`"EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x"`+`[color = red, shape = box]`+"\n",currentBlock.PrevHash[0:8],lastEpochBlock.Height,lastEpochBlock.MerklePatriciaRoot[0:8]))
			} else {
				FileConnections.WriteString(fmt.Sprintf(`"Hash : %x \n Height : %d" -> "Hash : %x \n Height : %d"`+"\n", currentBlock.PrevHash[0:8],currentBlock.Height-1,currentBlock.Hash[0:8],currentBlock.Height))
			}

			logger.Printf("Validated block: %vState:\n%v\n", currentBlock, getState())
			FileLogger.Printf("Validated block: %vState:\n%v\n", currentBlock, getState())
			broadcastBlock(currentBlock)
		} else {
			logger.Printf("Mined block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err.Error())
			FileLogger.Printf("Mined block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err.Error())
		}

		//if (prevBlockIsEpochBlock == true || FirstStartAfterEpoch==true) {
		//	//err := validateAfterEpoch(currentBlock, false) //here, block is written to closed storage and globalblockcount increased
		//	err := validate(currentBlock, false) //here, block is written to closed storage and globalblockcount increased
		//	if err == nil{
		//		stateTransition := protocol.NewStateTransition(storage.RelativeState,int(currentBlock.Height),storage.ThisShardID,currentBlock.Hash,
		//			currentBlock.ContractTxData,currentBlock.FundsTxData,currentBlock.ConfigTxData,currentBlock.StakeTxData)
		//		//logger.Printf("Broadcast state transition for height %d\n", currentBlock.Height)
		//		FileLogger.Printf("Broadcast state transition for height %d\n", currentBlock.Height)
		//		broadcastStateTransition(stateTransition)
		//		storage.OwnStateTransitionStash = append(storage.OwnStateTransitionStash,stateTransition)
		//
		//		FileLogger.Printf("Broadcast block for height %d\n", currentBlock.Height)
		//		broadcastBlock(currentBlock)
		//		FileConnections.WriteString(fmt.Sprintf(`"EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x" -> "Hash : %x \n Height : %d"`+"\n", currentBlock.PrevHash[0:8],lastEpochBlock.Height,lastEpochBlock.MerklePatriciaRoot[0:8],currentBlock.Hash[0:8],currentBlock.Height))
		//		FileConnections.WriteString(fmt.Sprintf(`"EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x"`+`[color = red, shape = box]`+"\n",currentBlock.PrevHash[0:8],lastEpochBlock.Height,lastEpochBlock.MerklePatriciaRoot[0:8]))
		//		logger.Printf("Validated block: %vState:\n%v\n", currentBlock, getState())
		//		FileLogger.Printf("Validated block: %vState:\n%v\n", currentBlock, getState())
		//	} else {
		//		logger.Printf("Mined block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err.Error())
		//		FileLogger.Printf("Mined block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err.Error())
		//	}
		//} else {
		//	err := validate(currentBlock, false) //here, block is written to closed storage and globalblockcount increased
		//	if err == nil {
		//		stateTransition := protocol.NewStateTransition(storage.RelativeState,int(currentBlock.Height),storage.ThisShardID,currentBlock.Hash,
		//			currentBlock.ContractTxData,currentBlock.FundsTxData,currentBlock.ConfigTxData,currentBlock.StakeTxData)
		//		//logger.Printf("Broadcast state transition for height %d\n", currentBlock.Height)
		//		FileLogger.Printf("Broadcast state transition for height %d\n", currentBlock.Height)
		//		broadcastStateTransition(stateTransition)
		//		storage.OwnStateTransitionStash = append(storage.OwnStateTransitionStash,stateTransition)
		//		FileLogger.Printf("Broadcast block for height %d\n", currentBlock.Height)
		//		broadcastBlock(currentBlock)
		//		logger.Printf("Validated block: %vState:\n%v\n", currentBlock, getState())
		//		FileLogger.Printf("Validated block: %vState:\n%v\n", currentBlock, getState())
		//		//FileConnections.WriteString(fmt.Sprintf("'%x' -> '%x'\n", currentBlock.PrevHash[0:15], currentBlock.Hash[0:15]))
		//		FileConnections.WriteString(fmt.Sprintf(`"Hash : %x \n Height : %d" -> "Hash : %x \n Height : %d"`+"\n", currentBlock.PrevHash[0:8],currentBlock.Height-1,currentBlock.Hash[0:8],currentBlock.Height))
		//	} else {
		//		logger.Printf("Mined block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err.Error())
		//		FileLogger.Printf("Mined block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err.Error())
		//	}
		//}
	}

	FirstStartAfterEpoch = false

}

func DetNumberOfShards() (numberOfShards int) {
	return int(math.Ceil(float64(GetValidatorsCount()) / float64(activeParameters.validators_per_shard)))
}

func AssignValidatorsToShards() map[[64]byte]int {

	/*This map denotes which validator is assigned to which shard index*/
	validatorShardAssignment := make(map[[64]byte]int)

	/*Fill validatorAssignedMap with the validators of the current state.
	The bool value indicates whether the validator has been assigned to a shard
	*/
	validatorSlices := make([][64]byte, 0)
	validatorAssignedMap := make(map[[64]byte]bool)
	for _, acc := range storage.State {
		if acc.IsStaking {
			validatorAssignedMap[acc.Address] = false
			validatorSlices = append(validatorSlices, acc.Address)
		}
	}

	/*Iterate over range of shards. At each index, select a random validators
	from the map above and set is bool 'assigned' to TRUE*/
	rand.Seed(time.Now().Unix())

	for j := 1; j <= int(activeParameters.validators_per_shard); j++ {
		for i := 1; i <= NumberOfShards; i++ {

			if len(validatorSlices) == 0 {
				return validatorShardAssignment
			}

			randomIndex := rand.Intn(len(validatorSlices))
			randomValidator := validatorSlices[randomIndex]

			//Assign validator to shard ID
			validatorShardAssignment[randomValidator] = i
			//Remove assigned validator from active list
			validatorSlices = removeValidator(validatorSlices, randomIndex)
		}
	}
	return validatorShardAssignment
}

func removeValidator(inputSlice [][64]byte, index int) [][64]byte {
	inputSlice[index] = inputSlice[len(inputSlice)-1]
	inputSlice = inputSlice[:len(inputSlice)-1]
	return inputSlice
}

func TxPayloadCointained(payloadSlice []*protocol.TransactionPayload, tp *protocol.TransactionPayload) bool {
	for _, v := range payloadSlice {
		if v == tp {
			return true
		}
	}
	return false
}

func removeBlock(blockSlice []*protocol.Block, bl *protocol.Block) []*protocol.Block {
	for i, v := range blockSlice {
		if v == bl {
			return append(blockSlice[:i], blockSlice[i+1:]...)
		}
	}
	return blockSlice
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}