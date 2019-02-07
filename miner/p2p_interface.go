package miner

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

//The code in this source file communicates with the p2p package via channels

//Constantly listen to incoming data from the network
func incomingData() {
	for {
		block := <-p2p.BlockIn
		processBlock(block)
	}
}

func incomingStateData(){
	for{
		stateTransition := <- p2p.StateTransitionIn
		processStateData(stateTransition)
	}
}

func incomingEpochData() {
	for {
		//receive Epoch Block
		epochBlock := <-p2p.EpochBlockIn
		processEpochBlock(epochBlock)
	}
}


func processEpochBlock(eb []byte) {
	var epochBlock *protocol.EpochBlock
	epochBlock = epochBlock.Decode(eb)

	if(storage.ReadClosedEpochBlock(epochBlock.Hash) != nil){
		logger.Printf("Received Epoch Block (%x) already in storage\n", epochBlock.Hash[0:8])
		FileLogger.Printf("Received Epoch Block (%x) already in storage\n", epochBlock.Hash[0:8])
		return
	} else {
		logger.Printf("Received Epoch Block: %v\n", epochBlock.String())
		FileLogger.Printf("Received Epoch Block: %v\n", epochBlock.String())
		ValidatorShardMap = epochBlock.ValMapping
		NumberOfShards = epochBlock.NofShards
		storage.ThisShardID = ValidatorShardMap.ValMapping[validatorAccAddress]
		lastEpochBlock = epochBlock
		storage.WriteClosedEpochBlock(epochBlock)
		broadcastEpochBlock(lastEpochBlock)
	}
}

func processStateData(payload []byte) {
	var stateTransition *protocol.StateTransition
	stateTransition = stateTransition.DecodeTransition(payload)

	if(lastEpochBlock != nil){
		if (stateTransition.ShardID != storage.ThisShardID && stateTransition.Height > int(lastEpochBlock.Height)){
			stateHash := stateTransition.HashTransition()
			if (storage.ReceivedStateStash.StateTransitionIncluded(stateHash) == false){
				FileLogger.Printf("Writing state to stash Shard ID: %v  VS my shard ID: %v - Height: %d - Hash: %x\n",stateTransition.ShardID,storage.ThisShardID,stateTransition.Height,stateHash[0:8])
				storage.ReceivedStateStash.Set(stateHash,stateTransition)
				FileLogger.Printf("Length state stash map: %d\n",len(storage.ReceivedStateStash.M))
				FileLogger.Printf("Length state stash keys: %d\n",len(storage.ReceivedStateStash.Keys))
				FileLogger.Printf("Redistributing state transition\n")
				broadcastStateTransition(stateTransition)
			} else {
				FileLogger.Printf("Received state transition already included: Shard ID: %v  VS my shard ID: %v - Height: %d - Hash: %x\n",stateTransition.ShardID,storage.ThisShardID,stateTransition.Height,stateHash[0:8])
				return
			}
		}
	}
}

func processBlock(payload []byte) {
	var block *protocol.Block
	block = block.Decode(payload)

	if block.ShardId == storage.ThisShardID && block.Height > lastEpochBlock.Height {

		FileLogger.Printf("Received block (%x) from own shard with height: %d \n", block.Hash[0:8],block.Height)

		//Block already confirmed and validated
		if storage.ReadClosedBlock(block.Hash) != nil {
			logger.Printf("Received block (%x) has already been validated.\n", block.Hash[0:8])
			FileLogger.Printf("Received block (%x) has already been validated.\n", block.Hash[0:8])
			return
		}

		//Append received Block to stash
		storage.WriteToReceivedStash(block)

		//Start validation process
		if(block.Height == lastEpochBlock.Height +1){
			err := validateAfterEpoch(block, false)
			if err == nil {
				logger.Printf("Received Validated block: %vState:\n%v\n", block, getState())
				FileLogger.Printf("Received Validated block: %vState:\n%v\n", block, getState())
				FileConnections.WriteString(fmt.Sprintf(`"EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x" -> "Hash : %x \n Height : %d"`+"\n", block.PrevHash[0:8],lastEpochBlock.Height,lastEpochBlock.MerklePatriciaRoot[0:8],block.Hash[0:8],block.Height))
				FileConnections.WriteString(fmt.Sprintf(`"EPOCH BLOCK: \n Hash : %x \n Height : %d \nMPT : %x"`+`[color = red, shape = box]`+"\n",block.PrevHash[0:8],lastEpochBlock.Height,lastEpochBlock.MerklePatriciaRoot[0:8]))
				broadcastBlock(block)
			} else {
				logger.Printf("Received block (%x) could not be validated: %v\n", block.Hash[0:8], err)
				FileLogger.Printf("Received block (%x) could not be validated: %v\n", block.Hash[0:8], err)
			}
		} else {
			err := validate(block, false)
			if err == nil {
				logger.Printf("Received Validated block: %vState:\n%v\n", block, getState())
				FileLogger.Printf("Received Validated block: %vState:\n%v\n", block, getState())
				FileConnections.WriteString(fmt.Sprintf(`"Hash : %x \n Height : %d" -> "Hash : %x \n Height : %d"`+"\n", block.PrevHash[0:8],block.Height-1,block.Hash[0:8],block.Height))
				broadcastBlock(block)
			} else {
				logger.Printf("Received block (%x) could not be validated: %v\n", block.Hash[0:8], err)
				FileLogger.Printf("Received block (%x) could not be validated: %v\n", block.Hash[0:8], err)
			}
		}
	} else {
		FileLogger.Printf("Writing block to stash Shard ID: %v  VS my shard ID: %v - Height: %d\n",block.ShardId,storage.ThisShardID,block.Height)
		//broadcastBlock(block)
		storage.ReceivedBlockStashFromOtherShards.Set(block.Hash,block)
		FileLogger.Printf("Length block stash: %d\n",len(storage.ReceivedBlockStashFromOtherShards.M))
		FileLogger.Printf("Length block stash: %d\n",len(storage.ReceivedBlockStashFromOtherShards.M))
	}
}

//p2p.BlockOut is a channel whose data get consumed by the p2p package
func broadcastBlock(block *protocol.Block) {
	p2p.BlockOut <- block.Encode()

	//Make a deep copy of the block (since it is a pointer and will be saved to db later).
	//Otherwise the block's bloom filter is initialized on the original block.
	var blockCopy = *block
	blockCopy.InitBloomFilter(append(storage.GetTxPubKeys(&blockCopy)))
	p2p.BlockHeaderOut <- blockCopy.EncodeHeader()
}

func broadcastStateTransition(st *protocol.StateTransition) {
	p2p.StateTransitionOut <- st.EncodeTransition()
}

func broadcastEpochBlock(epochBlock *protocol.EpochBlock) {
	p2p.EpochBlockOut <- epochBlock.Encode()
}


func broadcastVerifiedTxs(txs []*protocol.FundsTx) {
	var verifiedTxs [][]byte

	for _, tx := range txs {
		verifiedTxs = append(verifiedTxs, tx.Encode()[:])
	}

	p2p.VerifiedTxsOut <- protocol.Encode(verifiedTxs, protocol.FUNDSTX_SIZE)
}

func HashSliceContains(slice [][32]byte, hash [32]byte) bool {
	for _, a := range slice {
		if a == hash {
			return true
		}
	}
	return false
}

func IntSliceContains(slice []int, id int) bool {
	for _, a := range slice {
		if a == id {
			return true
		}
	}
	return false
}