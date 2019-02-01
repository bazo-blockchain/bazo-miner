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
	blockValidation         = &sync.Mutex{}
	payloadMap	            = &sync.Mutex{}
	lastShardMutex			= &sync.Mutex{}
	parameterSlice          []Parameters
	activeParameters        *Parameters
	uptodate                bool
	prevBlockIsEpochBlock   bool
	FirstStartAfterEpoch	bool
	slashingDict            = make(map[[64]byte]SlashingProof)
	validatorAccAddress     [64]byte
	ThisShardID             int // ID of the shard this validator is assigned to
	commPrivKey             *rsa.PrivateKey
	NumberOfShards          int
	ReceivedBlocksAtHeightX int //This counter is used to sync block heights among shards
	LastShardHashes         [][32]byte // This slice stores the hashes of the last blocks from the other shards, needed to create the next epoch block
	LastShardHashesMap 		= make(map[[32]byte][32]byte)
	ValidatorShardMap       *protocol.ValShardMapping // This map keeps track of the validator assignment to the shards; int: shard ID; [64]byte: validator address
	FileConnections   	       *os.File
	FileConnectionsLog         *os.File
	TransactionPayloadOut 	*protocol.TransactionPayload
	TransactionPayloadReceived 	[]*protocol.TransactionPayload
	TransactionPayloadReceivedMap 	= make(map[[32]byte]*protocol.TransactionPayload)
	TransactionPayloadIn 	[]*protocol.TransactionPayload
	processedTXPayloads		[]int //This slice keeps track of the tx payloads processed from a certain shard
	validatedTXCount		int
	validatedBlockCount		int
	blockStartTime			int64
	syncStartTime			int64
	blockEndTime			int64
	totalSyncTime			int64
	totalBlockTIme			int64
)

func InitFirstStart(wallet *ecdsa.PublicKey, commitment *rsa.PrivateKey) error {
	var err error
	FileConnections, err = os.OpenFile(fmt.Sprintf("hash-prevhash-%v.txt",strings.Split(p2p.Ipport, ":")[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	FileConnectionsLog, err = os.OpenFile(fmt.Sprintf("hlog-for-%v.txt",strings.Split(p2p.Ipport, ":")[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	logger = storage.InitLogger()

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
	firstValMapping := protocol.NewMapping()
	initialEpochBlock.ValMapping = firstValMapping
	//logger.Printf("Written Epoch Block: %v\n", initialEpochBlock.String())
	//FileConnectionsLog.WriteString(fmt.Sprintf("Written Epoch Block: %v\n", initialEpochBlock.String()))

	//FileConnections.WriteString(fmt.Sprintf("'GENESIS: %x' -> 'EPOCH BLOCK: %x'\n", [32]byte{}, initialEpochBlock.Hash[0:15]))
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
	//go incomingTxPayloadData()
	//go incomingValShardData()
	//go syncBlockHeight()

	//Since new validators only join after the currently running epoch ends, they do no need to download the whole shardchain history,
	//but can continue with their work after the next epoch block and directly set their state to the global state of the first received epoch block
	if(p2p.IsBootstrap()){
		initialBlock, err = initState() //From here on, every validator should have the same state representation
		if err != nil {
			return err
		}
		//FileConnections.WriteString(fmt.Sprintf("'%x' -> '%x'\n", initialBlock.PrevHash[0:15], initialBlock.Hash[0:15]))
		FileConnections.WriteString(fmt.Sprintf(`"Hash : %x \n Height : %d" -> "Hash : %x \n Height : %d"`+"\n", initialBlock.PrevHash[0:8],initialBlock.Height-1,initialBlock.Hash[0:8],initialBlock.Height))
		lastBlock = initialBlock
	} else {
		/*//Request last epoch block to check if I am already in the validator set, if so, then no need to wait for the next epoch
		lastEpochBlock, err = getLastEpochBlock()
		if err != nil {
			return err
		}
		if acc := lastEpochBlock.State[validatorAccAddress]; acc != nil{
			if(acc.IsStaking == true){
				storage.State = lastEpochBlock.State
			}
		}*/

		for{
			//Wait until I receive the last epoch block as well as the validator assignment
			// The global variables 'lastEpochBlock' and 'ValidatorShardMap' are being set when they are received by the network
			if(lastEpochBlock != nil && ValidatorShardMap != nil){
				storage.State = lastEpochBlock.State
				NumberOfShards = lastEpochBlock.NofShards
				ThisShardID = ValidatorShardMap.ValMapping[validatorAccAddress] //Save my ShardID
				FirstStartAfterEpoch = true
				epochMining(lastEpochBlock.Hash,lastEpochBlock.Height) //start mining based on the received Epoch Block
			}
		}
	}

	logger.Printf("Active config params:%v\n", activeParameters)
	FileConnectionsLog.WriteString(fmt.Sprintf("Active config params:%v\n", activeParameters))

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
		FileConnectionsLog.WriteString(fmt.Sprintf("Validator Shard Mapping:\n"))
		FileConnectionsLog.WriteString(fmt.Sprintf(validatorShardMapping.String()+"\n"))
		//broadcast the generated map to the other validators
		//broadcastValidatorShardMapping(ValidatorShardMap)
	}

	ThisShardID = ValidatorShardMap.ValMapping[validatorAccAddress]

	//logger.Printf("Entering epoch mining for the first time...")
	//FileConnectionsLog.WriteString(fmt.Sprintf("Entering epoch mining for the first time..."))
	epochMining(lastBlock.Hash, lastBlock.Height)

	//mining(initialBlock)

	return nil
}

func epochMining(hashPrevBlock [32]byte, heightPrevBlock uint32) {

	var epochBlock *protocol.EpochBlock

	/*constantly watch height of the lastblock variable for insertion of next epoch block*/
	for {
		//logger.Printf("Number of GoRoutines: %d\n",countGoRoutines())
		//FileConnectionsLog.WriteString(fmt.Sprintf("Number of GoRoutines: %d\n",countGoRoutines()))

		if(FirstStartAfterEpoch == true){ //Indicates that a validator newly joined Bazo after the current epoch, thus his 'lastBlock' variable is nil
			mining(hashPrevBlock,heightPrevBlock)
		}

		/*This is how it is done currently*/
		/*Wait until block heights of the shards are synched, i.e. until I received all the blocks with the height of my lastblock*/
		//logger.Printf("Before checking my blockStash for lastblock height: %d",lastBlock.Height)
		//FileConnectionsLog.WriteString(fmt.Sprintf("Before checking my blockStash for lastblock height: %d",lastBlock.Height))
		//for{
		//	if(NumberOfShards == 1 || storage.ReceivedBlockStash == nil){
		//		break
		//	}
		//
		//	if(protocol.CheckForHeight(storage.ReceivedBlockStash,lastBlock.Height) == NumberOfShards-1){
		//		TransactionPayloadIn = protocol.ReturnTxPayloadForHeight(storage.ReceivedBlockStash,lastBlock.Height)
		//		//Sync state with the other shards
		//		err := updateGlobalState(TransactionPayloadIn)
		//		if err != nil {
		//			//return
		//			logger.Printf("error in updating state: %s",err.Error())
		//			FileConnectionsLog.WriteString(fmt.Sprintf("error in updating state: %s",err.Error()))
		//		}
		//		//logger.Printf("After synchronizing global state for height: %d",lastBlock.Height)
		//		//FileConnectionsLog.WriteString(fmt.Sprintf("After synchronizing global state for height: %d",lastBlock.Height))
		//		break
		//	}
		//}
		//logger.Printf("After checking my blockStash for lastblock height: %d",lastBlock.Height)
		//FileConnectionsLog.WriteString(fmt.Sprintf("After checking my blockStash for lastblock height: %d",lastBlock.Height))
		//
		//logger.Printf("Within overall for-loop Height: %d - MyShardID: %d\n",lastBlock.Height,ThisShardID)
		//FileConnectionsLog.WriteString(fmt.Sprintf("Within overall for-loop Height: %d - MyShardID: %d\n",lastBlock.Height,ThisShardID))

		/*This is how it is done with state transition*/
		logger.Printf("Before checking my state stash for lastblock height: %d\n",lastBlock.Height)
		FileConnectionsLog.WriteString(fmt.Sprintf("Before checking my state stash for lastblock height: %d\n",lastBlock.Height))
		syncStartTime = time.Now().Unix()
		for{
			if(NumberOfShards == 1){
				break
			}

			if(protocol.CheckForHeightStateTransition(storage.ReceivedStateStash,lastBlock.Height) == NumberOfShards-1 &&
				protocol.CheckForHeight(storage.ReceivedBlockStash,lastBlock.Height) == NumberOfShards-1){
				stateTransitionSlice := protocol.ReturnStateTransitionForHeight(storage.ReceivedStateStash,lastBlock.Height)
				//Apply relative state transitions from other shards
				for _, st := range stateTransitionSlice {
					storage.State = storage.ApplyRelativeState(storage.State,st.RelativeStateChange)
				}

				//Delete transactions from mempool, which were validated by the other shards
				//TransactionPayloadIn = protocol.ReturnTxPayloadForHeight(storage.ReceivedBlockStash,lastBlock.Height)
				//DeleteTransactionFromMempool(TransactionPayloadIn)
				break
			}
		}
		logger.Printf("After checking my state stash for lastblock height: %d\n",lastBlock.Height)
		FileConnectionsLog.WriteString(fmt.Sprintf("After checking my state stash for lastblock height: %d\n",lastBlock.Height))

		var syncEndTime = time.Now().Unix()
		var syncDuration = syncEndTime - syncStartTime
		totalSyncTime += syncDuration
		FileConnectionsLog.WriteString(fmt.Sprintf("Synchronisation duration for lastblock height: %d - %d seconds\n",lastBlock.Height,syncDuration))
		FileConnectionsLog.WriteString(fmt.Sprintf("Total Synchronisation duration for lastblock height: %d - %d seconds\n",lastBlock.Height,totalSyncTime))

		prevBlockIsEpochBlock = false // this boolean indicates whether the previous block is an epoch block

		//TODO: @Kürsat: lol, remove if clause
		if (true) {
			if (lastBlock.Height == uint32(lastEpochBlock.Height) + uint32(activeParameters.epoch_length)) {
				//The variable 'lastblock' is one before the next epoch block, thus with the lastblock, crete next epoch block

				epochBlock = protocol.NewEpochBlock([][32]byte{lastBlock.Hash}, lastBlock.Height+1)

				////Get hashes of last shard blocks from other miners and include them in the next epoch block
				//logger.Printf("Before getting lastshard hashes for Epoch Height: %d\n",epochBlock.Height)
				//FileConnectionsLog.WriteString(fmt.Sprintf("Before getting lastshard hashes for Epoch Height: %d\n",epochBlock.Height))
				//logger.Printf("Size of lastshard hashes: %d for Epoch Height: %d\n",len(LastShardHashesMap),epochBlock.Height)
				//FileConnectionsLog.WriteString(fmt.Sprintf("Size of lastshard hashes: %d for Epoch Height: %d\n",len(LastShardHashesMap),epochBlock.Height))
				//for{
				//	/*Abort mining epoch block if one is received from the network*/
				//	if(lastEpochBlock.Height >= lastBlock.Height){
				//		logger.Printf("Abort creating epoch block, another miner has successfully mined one in the meantime\n")
				//		FileConnectionsLog.WriteString(fmt.Sprintf("Abort creating epoch block, another miner has successfully mined one in the meantime\n"))
				//		prevBlockIsEpochBlock = true
				//		mining(lastEpochBlock.Hash,lastEpochBlock.Height)
				//	}
				//	if(NumberOfShards == 1){
				//		break
				//	} else if(len(LastShardHashesMap) == NumberOfShards - 1 && NumberOfShards != 1){
				//		lastShardMutex.Lock()
				//		logger.Printf("Appending last shard hashes to the epoch block\n")
				//		FileConnectionsLog.WriteString(fmt.Sprintf("Appending last shard hashes to the epoch block\n"))
				//		for key, _ := range LastShardHashesMap {
				//			epochBlock.PrevShardHashes = append(epochBlock.PrevShardHashes,key)
				//		}
				//		lastShardMutex.Unlock()
				//
				//		//epochBlock.PrevShardHashes = append(epochBlock.PrevShardHashes, LastShardHashes...)
				//		LastShardHashesMap = nil // empty the map
				//		LastShardHashesMap = make(map[[32]byte][32]byte)
				//		break
				//	}
				//}
				//
				//logger.Printf("After getting lastshard hashes for Epoch Height: %d\n",epochBlock.Height)
				//FileConnectionsLog.WriteString(fmt.Sprintf("After getting lastshard hashes for Epoch Height: %d\n",epochBlock.Height))
				//logger.Printf("before finalizeEpochBlock() ---- Height: %d\n",epochBlock.Height)
				//FileConnectionsLog.WriteString(fmt.Sprintf("before finalizeEpochBlock() ---- Height: %d\n",epochBlock.Height))

				if(NumberOfShards != 1 || storage.ReceivedBlockStash != nil){
					/*Extract the hashes of the last block of the other shards, needed to create the epoch block*/
					LastShardHashes = protocol.ReturnHashesForHeight(storage.ReceivedBlockStash,lastBlock.Height)
					epochBlock.PrevShardHashes = append(epochBlock.PrevShardHashes,LastShardHashes...)
				}

				//logger.Printf("before finalizeEpochBlock() ---- Height: %d\n",epochBlock.Height)
				//FileConnectionsLog.WriteString(fmt.Sprintf("before finalizeEpochBlock() ---- Height: %d\n",epochBlock.Height))

				//Include validator-shard mapping in the epoch block
				err := finalizeEpochBlock(epochBlock) //in case another epoch block was mined in the meantime, abort PoS here

				//logger.Printf("after finalizeEpochBlock() ---- Height: %d\n",epochBlock.Height)
				//FileConnectionsLog.WriteString(fmt.Sprintf("after finalizeEpochBlock() ---- Height: %d\n",epochBlock.Height))
				if err != nil {
					logger.Printf("%v\n", err)
					FileConnectionsLog.WriteString(fmt.Sprintf("%v\n", err))
				} else {
					logger.Printf("EPOCH BLOCK mined (%x)\n", epochBlock.Hash[0:8])
					FileConnectionsLog.WriteString(fmt.Sprintf("EPOCH BLOCK mined (%x)\n", epochBlock.Hash[0:8]))
				}

				if err == nil {
					/*epochBlock.Hash = epochBlock.HashEpochBlock()
					epochBlock.State = storage.State*/

					//broadcast the generated map to the other validators
					/*broadcastValidatorShardMapping(ValidatorShardMap)*/
					//logger.Printf("Broadcast epoch block (%x)\n", epochBlock.Hash[0:8])
					//FileConnectionsLog.WriteString(fmt.Sprintf("Broadcast epoch block (%x)\n", epochBlock.Hash[0:8]))
					broadcastEpochBlock(epochBlock)
					storage.WriteClosedEpochBlock(epochBlock)
					storage.DeleteAllLastClosedEpochBlock()
					storage.WriteLastClosedEpochBlock(epochBlock)
					lastEpochBlock = epochBlock

					//storage.WriteValidatorMapping(ValidatorShardMap)
					logger.Printf("Created Validator Shard Mapping :\n")
					logger.Printf(ValidatorShardMap.String())
					logger.Printf("Inserting EPOCH BLOCK: %v\n", epochBlock.String())
					FileConnectionsLog.WriteString(fmt.Sprintf("Created Validator Shard Mapping :\n"))
					FileConnectionsLog.WriteString(fmt.Sprintf(ValidatorShardMap.String()+"\n"))
					FileConnectionsLog.WriteString(fmt.Sprintf("Inserting EPOCH BLOCK: %v\n", epochBlock.String()))

					for _, prevHash := range epochBlock.PrevShardHashes {
						//FileConnections.WriteString(fmt.Sprintf("'%x' -> 'EPOCH BLOCK: %x'\n", prevHash[0:15], epochBlock.Hash[0:15]))
						FileConnections.WriteString(fmt.Sprintf(`"Hash : %x \n Height : %d" -> "EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x"`+"\n", prevHash[0:8],epochBlock.Height-1,epochBlock.Hash[0:8],epochBlock.Height,epochBlock.MerklePatriciaRoot[0:8]))
						FileConnections.WriteString(fmt.Sprintf(`"EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x"`+`[color = red, shape = box]`+"\n",epochBlock.Hash[0:8],epochBlock.Height,epochBlock.MerklePatriciaRoot[0:8]))
					}
				}

				prevBlockIsEpochBlock = true

				firstEpochOver = true

				/*Wait until new validator-shard mapping is either created by me, or received from the network, then continue mining*/
				/*for{
					if(ValidatorShardMap.EpochHeight >= int(lastEpochBlock.Height)){
						break
					}
				}*/

				mining(lastEpochBlock.Hash, lastEpochBlock.Height)
			} else {
				mining(lastBlock.Hash, lastBlock.Height)
			}
		}
		//ReceivedBlocksAtHeightX = 0 // reset counter of received blocks from other shards
	}
}

//Mining is a constant process, trying to come up with a successful PoW.
func mining(hashPrevBlock [32]byte, heightPrevBlock uint32) {

	blockStartTime = time.Now().Unix()
	currentBlock := newBlock(hashPrevBlock, [crypto.COMM_PROOF_LENGTH]byte{}, heightPrevBlock+1)

	/*Set shard identifier in block*/
	currentBlock.ShardId = ThisShardID
	blockBeingProcessed = currentBlock

	logger.Printf("blockBeingProcessed Height: %v - MyShardID: %d\n",currentBlock.Height,ThisShardID)
	FileConnectionsLog.WriteString(fmt.Sprintf("blockBeingProcessed Height: %v - MyShardID: %d\n",currentBlock.Height,ThisShardID))
	//This is the same mutex that is claimed at the beginning of a block validation. The reason we do this is
	//that before start mining a new block we empty the mempool which contains tx data that is likely to be
	//validated with block validation, so we wait in order to not work on tx data that is already validated
	//when we finish the block.
	blockValidation.Lock()
	//FileConnectionsLog.WriteString(fmt.Sprintf("before preparing Block Height: %v\n",currentBlock.Height))
	//logger.Printf("blockBeingProcessed Height: %v - MyShardID: %d\n",currentBlock.Height,ThisShardID)
	prepareBlock(currentBlock) // In this step, filter tx from mem pool to check if they belong to my shard
	//FileConnectionsLog.WriteString(fmt.Sprintf("After preparing Block Height: %v\n",currentBlock.Height))
	//logger.Printf("blockBeingProcessed Height: %v - MyShardID: %d\n",currentBlock.Height,ThisShardID)
	blockValidation.Unlock()


	//Fill global variable TransactionPayloadOut to be broadcasted to the other shards
	/*TransactionPayloadOut = protocol.NewTransactionPayload(ThisShardID, int(currentBlock.Height),nil,nil,nil,nil)
	TransactionPayloadOut.ContractTxData = currentBlock.ContractTxData
	TransactionPayloadOut.FundsTxData = currentBlock.FundsTxData
	TransactionPayloadOut.ConfigTxData = currentBlock.ConfigTxData
	TransactionPayloadOut.StakeTxData = currentBlock.StakeTxData
	logger.Printf("---- Sending TX Payload ----: Height: %d -- Shard: %d \n",currentBlock.Height,ThisShardID)
	FileConnectionsLog.WriteString(fmt.Sprintf("---- Sending TX Payload ----: Height: %d -- Shard: %d \n",currentBlock.Height,ThisShardID))
	//logger.Printf(TransactionPayloadOut.StringPayload())
	broadcastTxPayload()*/

	_, err := storage.ReadAccount(validatorAccAddress)
	if err != nil {
		logger.Printf("%v\n", err)
		FileConnectionsLog.WriteString(fmt.Sprintf("%v\n", err))
		time.Sleep(10 * time.Second)
		return
	}

	//logger.Printf("---- Before finalizeBlock() ---- Height: %d\n",currentBlock.Height)
	//FileConnectionsLog.WriteString(fmt.Sprintf("---- Before finalizeBlock() ---- Height: %d\n",currentBlock.Height))

	err = finalizeBlock(currentBlock) //in case another block was mined in the meantime, abort PoS here

	//logger.Printf("---- After finalizeBlock() ---- Height: %d\n",currentBlock.Height)
	//FileConnectionsLog.WriteString(fmt.Sprintf("---- After finalizeBlock() ---- Height: %d\n",currentBlock.Height))

	if err != nil {
		logger.Printf("%v\n", err)
		FileConnectionsLog.WriteString(fmt.Sprintf("%v\n", err))
	} else {
		logger.Printf("Block mined (%x)\n", currentBlock.Hash[0:8])
		FileConnectionsLog.WriteString(fmt.Sprintf("Block mined (%x)\n", currentBlock.Hash[0:8]))
	}

	if err == nil {
		/*logger.Printf("Sending Block Height: %d", currentBlock.Height)
		FileConnectionsLog.WriteString(fmt.Sprintf("Sending Block Height: %d", currentBlock.Height))*/
		if (prevBlockIsEpochBlock == true || FirstStartAfterEpoch==true) {
			err := validateAfterEpoch(currentBlock, false) //here, block is written to closed storage and globalblockcount increased
			if err == nil{
				stateTransition := protocol.NewStateTransition(storage.RelativeState,int(currentBlock.Height),ThisShardID)
				//logger.Printf("Broadcast state transition for height %d\n", currentBlock.Height)
				//FileConnectionsLog.WriteString(fmt.Sprintf("Broadcast state transition for height %d\n", currentBlock.Height))
				broadcastStateTransition(stateTransition)
				broadcastBlock(currentBlock)
				FileConnections.WriteString(fmt.Sprintf(`"EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x" -> "Hash : %x \n Height : %d"`+"\n", currentBlock.PrevHash[0:8],lastEpochBlock.Height,lastEpochBlock.MerklePatriciaRoot[0:8],currentBlock.Hash[0:8],currentBlock.Height))
				FileConnections.WriteString(fmt.Sprintf(`"EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x"`+`[color = red, shape = box]`+"\n",currentBlock.PrevHash[0:8],lastEpochBlock.Height,lastEpochBlock.MerklePatriciaRoot[0:8]))
				logger.Printf("Validated block: %vState:\n%v\n", currentBlock, getState())
				FileConnectionsLog.WriteString(fmt.Sprintf("Validated block: %vState:\n%v\n", currentBlock, getState()))
			} else {
				logger.Printf("Mined block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err.Error())
				FileConnectionsLog.WriteString(fmt.Sprintf("Mined block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err.Error()))
			}
			//
			//blockDataMap := make(map[[32]byte]blockData)
			//contractTxs, fundsTxs, configTxs, stakeTxs, err := preValidate(currentBlock, false)
			//
			//if (err == nil) {
			//	//FileConnections.WriteString(fmt.Sprintf("'EPOCH BLOCK: %x' -> '%x'\n", currentBlock.PrevHash[0:15], currentBlock.Hash[0:15]))
			//	FileConnections.WriteString(fmt.Sprintf(`"EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x" -> "Hash : %x \n Height : %d"`+"\n", currentBlock.PrevHash[0:8],lastEpochBlock.Height,lastEpochBlock.MerklePatriciaRoot[0:8],currentBlock.Hash[0:8],currentBlock.Height))
			//	FileConnections.WriteString(fmt.Sprintf(`"EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x"`+`[color = red, shape = box]`+"\n",currentBlock.PrevHash[0:8],lastEpochBlock.Height,lastEpochBlock.MerklePatriciaRoot[0:8]))
			//	blockDataMap[currentBlock.Hash] = blockData{contractTxs, fundsTxs, configTxs, stakeTxs, currentBlock}
			//	if err := validateState(blockDataMap[currentBlock.Hash]); err != nil {
			//		logger.Printf("ERROR in validating State of Block: %vState:\n%v\n", currentBlock, getState())
			//		FileConnectionsLog.WriteString(fmt.Sprintf("ERROR in validating State of Block: %vState:\n%v\n", currentBlock, getState()))
			//		return
			//	}
			//	postValidate(blockDataMap[currentBlock.Hash], false)
			//	logger.Printf("Validated block: %vState:\n%v\n", currentBlock, getState())
			//	FileConnectionsLog.WriteString(fmt.Sprintf("Validated block: %vState:\n%v\n", currentBlock, getState()))
			//}
		} else {
			err := validate(currentBlock, false) //here, block is written to closed storage and globalblockcount increased
			if err == nil {
				stateTransition := protocol.NewStateTransition(storage.RelativeState,int(currentBlock.Height),ThisShardID)
				//logger.Printf("Broadcast state transition for height %d\n", currentBlock.Height)
				//FileConnectionsLog.WriteString(fmt.Sprintf("Broadcast state transition for height %d\n", currentBlock.Height))
				broadcastStateTransition(stateTransition)
				broadcastBlock(currentBlock)
				logger.Printf("Validated block: %vState:\n%v\n", currentBlock, getState())
				FileConnectionsLog.WriteString(fmt.Sprintf("Validated block: %vState:\n%v\n", currentBlock, getState()))
				//FileConnections.WriteString(fmt.Sprintf("'%x' -> '%x'\n", currentBlock.PrevHash[0:15], currentBlock.Hash[0:15]))
				FileConnections.WriteString(fmt.Sprintf(`"Hash : %x \n Height : %d" -> "Hash : %x \n Height : %d"`+"\n", currentBlock.PrevHash[0:8],currentBlock.Height-1,currentBlock.Hash[0:8],currentBlock.Height))
			} else {
				logger.Printf("Mined block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err.Error())
				FileConnectionsLog.WriteString(fmt.Sprintf("Mined block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err.Error()))
			}
		}
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
