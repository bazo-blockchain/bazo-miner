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
	ValidatorShardMap       *protocol.ValShardMapping // This map keeps track of the validator assignment to the shards; int: shard ID; [64]byte: validator address
	FileConnections         *os.File
	TransactionPayloadOut 	*protocol.TransactionPayload
	TransactionPayloadIn 	[]*protocol.TransactionPayload
	processedTXPayloads		[]int //This slice keeps track of the tx payloads processed from a certain shard
)

func InitFirstStart(wallet *ecdsa.PublicKey, commitment *rsa.PrivateKey) error {
	var err error
	FileConnections, err = os.OpenFile(fmt.Sprintf("hash-prevhash-%v.txt",strings.Split(p2p.Ipport, ":")[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	logger = storage.InitLogger()

	rootAddress := crypto.GetAddressFromPubKey(wallet)

	var rootCommitment [crypto.COMM_KEY_LENGTH]byte
	copy(rootCommitment[:], commitment.N.Bytes())

	genesis := protocol.NewGenesis(rootAddress, rootCommitment)
	storage.WriteGenesis(&genesis)

	logger.Printf("Written Genesis Block: %v\n", genesis.String())

	/*Write First Epoch block chained to the genesis block*/
	initialEpochBlock := protocol.NewEpochBlock([][32]byte{genesis.Hash()}, 0)
	initialEpochBlock.Hash = initialEpochBlock.HashEpochBlock()
	FirstEpochBlock = initialEpochBlock
	initialEpochBlock.State = storage.State
	storage.WriteFirstEpochBlock(initialEpochBlock)
	logger.Printf("Written Epoch Block: %v\n", initialEpochBlock.String())

	FileConnections.WriteString(fmt.Sprintf("'GENESIS: %x' -> 'EPOCH BLOCK: %x'\n", [32]byte{}, initialEpochBlock.Hash[0:15]))

	return Init(wallet, commitment)
}

//Miner entry point
func Init(wallet *ecdsa.PublicKey, commitment *rsa.PrivateKey) error {
	//this bool indicates whether the first epoch is over. Only in the first epoch, the bootstrapping node is assigning the
	//validators to the shards and broadcasts this assignment to the other miners
	firstEpochOver = false

	FileConnections, _ = os.OpenFile(fmt.Sprintf("hash-prevhash-%v.txt",strings.Split(p2p.Ipport, ":")[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

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

	//Since new validators only join after the currently running epoch ends, they do no need to download the whole shardchain history,
	//but can continue with their work after the next epoch block and directly set their state to the global state of the next epoch block
	if(p2p.IsBootstrap()){
		initialBlock, err = initState() //From here on, every validator should have the same state representation
		if err != nil {
			return err
		}
		FileConnections.WriteString(fmt.Sprintf("'%x' -> '%x'\n", initialBlock.PrevHash[0:15], initialBlock.Hash[0:15]))
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
				NumberOfShards = DetNumberOfShards()
				ThisShardID = ValidatorShardMap.ValMapping[validatorAccAddress] //Save my ShardID
				FirstStartAfterEpoch = true
				epochMining(lastEpochBlock.Hash,lastEpochBlock.Height) //start mining based on the received Epoch Block
			}

			/*//Wait until the next epoch block and validator-shard mapping is broadcasted, then continue with mining
			encodedEpochBlockToStart := <-p2p.EpochBlockIn
			epochBlockToStart = epochBlockToStart.Decode(encodedEpochBlockToStart)
			logger.Printf("Received first Epoch Block: %v\n", epochBlockToStart.String())

			encodedValidatorMappping := <-p2p.ValidatorShardMapChanIn
			validatorMappingToStart = validatorMappingToStart.Decode(encodedValidatorMappping)
			logger.Printf("Received first Validaor Shard Mapping: %v\n", validatorMappingToStart.String())*/
			/*
			if (epochBlockToStart != nil && validatorMappingToStart != nil){
				storage.State = epochBlockToStart.State
				break
			}*/
		}
	}

	logger.Printf("Active config params:%v", activeParameters)

	/*Sharding Utilities*/
	NumberOfShards = DetNumberOfShards()

	/*First validator assignment is done by the bootstrapping node, the others will be done based on POS at the end of each epoch*/
	if (p2p.IsBootstrap()) {
		var validatorShardMapping = protocol.NewMapping()
		validatorShardMapping.ValMapping = AssignValidatorsToShards()
		ValidatorShardMap = validatorShardMapping
		storage.WriteValidatorMapping(ValidatorShardMap)
		logger.Printf("Validator Shard Mapping:\n")
		logger.Printf(validatorShardMapping.String())

		//broadcast the generated map to the other validators
		broadcastValidatorShardMapping(ValidatorShardMap)
	}
	/*else {
		//request the mapping from bootstrapping node
		p2p.ValidatorShardMapRequest()
		//Blocking wait
		select {
		case encodedMapping := <-p2p.ValidatorShardMapReq:
			var rvdMapping = protocol.NewMapping()
			rvdMapping = rvdMapping.Decode(encodedMapping)
			ValidatorShardMap = rvdMapping
			storage.WriteValidatorMapping(ValidatorShardMap)
			logger.Printf("Validator Shard Mapping:\n")
			logger.Printf(rvdMapping.String())

			//Limit waiting time to BLOCKFETCH_TIMEOUT seconds before aborting.
		case <-time.After(BLOCKFETCH_TIMEOUT * time.Second):
			return errors.New("mapping fetch timeout")
		}
	}*/

	ThisShardID = ValidatorShardMap.ValMapping[validatorAccAddress]

	epochMining(lastBlock.Hash, lastBlock.Height)

	//mining(initialBlock)

	return nil
}

func epochMining(hashPrevBlock [32]byte, heightPrevBlock uint32) {

	var epochBlock *protocol.EpochBlock
	//currentBlock := newBlock(initialBlock.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, initialBlock.Height + 1)

	/*constantly watch global blockcount for insertion of next epoch block*/
	for {
		prevBlockIsEpochBlock = false // this boolean indicates whether the previous block is an epoch block

		if(FirstStartAfterEpoch == true){ //Indicates that a validator newly joined Bazo after the current epoch, thus his 'lastBlock' variable is nil
			mining(hashPrevBlock,heightPrevBlock)
		}

		//shard height synced with network, now mine next block
		if (ReceivedBlocksAtHeightX == NumberOfShards-1) {
		//if (true) {
			if (int(lastBlock.Height) == int(lastEpochBlock.Height) + activeParameters.epoch_length) {
				//The variable 'lastblock' is one before the next epoch block, thus with the lastblock, crete next epoch block

				epochBlock = protocol.NewEpochBlock([][32]byte{lastBlock.Hash}, lastBlock.Height+1)
				//Get hashes of last shard blocks from other miners and include them in the next epoch block
				if len(LastShardHashes) == NumberOfShards-1 && NumberOfShards != 1 {
					epochBlock.PrevShardHashes = append(epochBlock.PrevShardHashes, LastShardHashes...)
					LastShardHashes = nil // empty the slice
				}

				err := finalizeEpochBlock(epochBlock) //in case another epoch block was mined in the meantime, abort PoS here
				if err != nil {
					logger.Printf("%v\n", err)
				} else {
					logger.Printf("EPOCH BLOCK mined (%x)\n", epochBlock.Hash[0:8])
				}

				if err == nil {
					epochBlock.Hash = epochBlock.HashEpochBlock()
					epochBlock.State = storage.State
					broadcastEpochBlock(epochBlock)
					storage.WriteClosedEpochBlock(epochBlock)
					storage.DeleteAllLastClosedEpochBlock()
					storage.WriteLastClosedEpochBlock(epochBlock)
					lastEpochBlock = epochBlock
					logger.Printf("Inserting EPOCH BLOCK: %v\n", epochBlock.String())

					for _, prevHash := range epochBlock.PrevShardHashes {
						FileConnections.WriteString(fmt.Sprintf("'%x' -> 'EPOCH BLOCK: %x'\n", prevHash[0:15], epochBlock.Hash[0:15]))
					}

					/*Determine new number of shards needed based on current state*/
					NumberOfShards = DetNumberOfShards()

					//generate new validator mapping and broadcast mapping to validators
					ValidatorShardMap.ValMapping = AssignValidatorsToShards()
					ThisShardID = ValidatorShardMap.ValMapping[validatorAccAddress]
					storage.WriteValidatorMapping(ValidatorShardMap)
					logger.Printf("Validator Shard Mapping:\n")
					logger.Printf(ValidatorShardMap.String())

					//broadcast the generated map to the other validators
					broadcastValidatorShardMapping(ValidatorShardMap)
				}

				prevBlockIsEpochBlock = true

				firstEpochOver = true

				mining(epochBlock.Hash, epochBlock.Height)
			} else {
				mining(lastBlock.Hash, lastBlock.Height)
			}
		}
		ReceivedBlocksAtHeightX = 0 // reset counter of received blocks from other shards
	}
}

//Mining is a constant process, trying to come up with a successful PoW.
func mining(hashPrevBlock [32]byte, heightPrevBlock uint32) {
	currentBlock := newBlock(hashPrevBlock, [crypto.COMM_PROOF_LENGTH]byte{}, heightPrevBlock+1)

	/*Set shard identifier in block*/
	currentBlock.ShardId = ThisShardID

	//This is the same mutex that is claimed at the beginning of a block validation. The reason we do this is
	//that before start mining a new block we empty the mempool which contains tx data that is likely to be
	//validated with block validation, so we wait in order to not work on tx data that is already validated
	//when we finish the block.
	blockValidation.Lock()
	//nextBlock := newBlock(lastBlock.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, lastBlock.Height+1)
	//currentBlock = nextBlock
	prepareBlock(currentBlock) // In this step, filter tx from mem pool to check if they belong to my shard
	blockValidation.Unlock()

	blockBeingProcessed = currentBlock

	//Fill global variable transactionPayload to be broadcasted to the other shards
	TransactionPayloadOut = protocol.NewTransactionPayload(ThisShardID, int(currentBlock.Height),nil,nil,nil,nil)
	TransactionPayloadOut.ContractTxData = currentBlock.ContractTxData
	TransactionPayloadOut.FundsTxData = currentBlock.FundsTxData
	TransactionPayloadOut.ConfigTxData = currentBlock.ConfigTxData
	TransactionPayloadOut.StakeTxData = currentBlock.StakeTxData
	broadcastTxPayload()

	_, err := storage.ReadAccount(validatorAccAddress)
	if err != nil {
		logger.Printf("%v\n", err)
		time.Sleep(10 * time.Second)
		return
	}

	err = finalizeBlock(currentBlock) //in case another block was mined in the meantime, abort PoS here

	if err != nil {
		logger.Printf("%v\n", err)
	} else {
		logger.Printf("Block mined (%x)\n", currentBlock.Hash[0:8])
	}

	if err == nil {
		broadcastBlock(currentBlock)
		logger.Printf("---- TX PAYLOAD ----")
		logger.Printf(TransactionPayloadOut.StringPayload())

		if (prevBlockIsEpochBlock == true || FirstStartAfterEpoch==true) {
			blockDataMap := make(map[[32]byte]blockData)
			contractTxs, fundsTxs, configTxs, stakeTxs, err := preValidate(currentBlock, false)

			if (err == nil) {
				FileConnections.WriteString(fmt.Sprintf("'EPOCH BLOCK: %x' -> '%x'\n", currentBlock.PrevHash[0:15], currentBlock.Hash[0:15]))
				blockDataMap[currentBlock.Hash] = blockData{contractTxs, fundsTxs, configTxs, stakeTxs, currentBlock}
				// validateState() and check for error, if not, then continue with postValidate(...)
				if err := validateState(blockDataMap[currentBlock.Hash]); err != nil {
					logger.Printf("ERROR in validating State of Block: %vState:\n%v", currentBlock, getState())
					return
				}
				postValidate(blockDataMap[currentBlock.Hash], false)
				logger.Printf("Validated block: %vState:\n%v", currentBlock, getState())
			}

		} else {
			err := validate(currentBlock, false) //here, block is written to closed storage and globalblockcount increased
			if err == nil {
				logger.Printf("Validated block: %vState:\n%v", currentBlock, getState())
				FileConnections.WriteString(fmt.Sprintf("'%x' -> '%x'\n", currentBlock.PrevHash[0:15], currentBlock.Hash[0:15]))
			} else {
				logger.Printf("Received block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err)
			}
		}
	}
	FirstStartAfterEpoch = false

	/*for {
		_, err := storage.ReadAccount(validatorAccAddress)
		if err != nil {
			logger.Printf("%v\n", err)
			time.Sleep(10 * time.Second)
			continue
		}

		err = finalizeBlock(currentBlock)
		if err != nil {
			logger.Printf("%v\n", err)
		} else {
			logger.Printf("Block mined (%x)\n", currentBlock.Hash[0:8])
		}

		if err == nil {
			broadcastBlock(currentBlock)
			err := validate(currentBlock, false) //here, block is written to closed storage
			if err == nil {
				logger.Printf("Validated block: %vState:\n%v", currentBlock, getState())
				FileConnections.WriteString(fmt.Sprintf("'%x' -> '%x'\n",currentBlock.PrevHash[0:15],currentBlock.Hash[0:15]))
			} else {
				logger.Printf("Received block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err)
			}
		}

		//This is the same mutex that is claimed at the beginning of a block validation. The reason we do this is
		//that before start mining a new block we empty the mempool which contains tx data that is likely to be
		//validated with block validation, so we wait in order to not work on tx data that is already validated
		//when we finish the block.
		blockValidation.Lock()
		nextBlock := newBlock(lastBlock.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, lastBlock.Height+1)
		currentBlock = nextBlock
		prepareBlock(currentBlock) // In this step, filter tx from mem pool to check if they belong to my shard
		blockValidation.Unlock()
	}*/
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

	for j := 1; j <= activeParameters.validators_per_shard; j++ {
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
