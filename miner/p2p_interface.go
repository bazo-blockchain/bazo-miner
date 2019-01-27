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

func incomingEpochData() {
	for {
		//receive Epoch Block
		epochBlock := <-p2p.EpochBlockIn
		processEpochBlock(epochBlock)
	}
}

func incomingValShardData() {
	for {
		//check validator assignment
		validatorMapping := <- p2p.ValidatorShardMapChanIn
		processReceivedValidatorMapping(validatorMapping)
	}
}

func incomingTxPayloadData() {
	for {
		//TX payload in
		txPayload := <- p2p.TxPayloadIn
		processPayload(txPayload)
	}
}

func syncBlockHeight(){
	for{
		for _, bl := range blocksReceived{
			if(blockBeingProcessed != nil){
				if bl.Height == blockBeingProcessed.Height{
					ReceivedBlocksAtHeightX = ReceivedBlocksAtHeightX + 1
					logger.Printf("Increased ReceivedBlocksAtHeightX: %d",ReceivedBlocksAtHeightX)
					removeBlock(blocksReceived, bl)
				}
			}
		}
	}
}

func processReceivedValidatorMapping(vm []byte) {
	var valMapping *protocol.ValShardMapping
	valMapping = valMapping.Decode(vm)
	ValidatorShardMap = valMapping
	ThisShardID = ValidatorShardMap.ValMapping[validatorAccAddress]
	logger.Printf("Received Validator Shard Mapping:\n")
	logger.Printf(ValidatorShardMap.String())
	FileConnectionsLog.WriteString(fmt.Sprintf("Received Validator Shard Mapping:\n"))
	FileConnectionsLog.WriteString(fmt.Sprintf(ValidatorShardMap.String()+"\n"))
	//broadcastValidatorShardMapping(valMapping)
}
func processPayload(payloadByte []byte) {
	var payload *protocol.TransactionPayload
	payload = payload.DecodePayload(payloadByte)

	payloadMap.Lock()
	TransactionPayloadReceivedMap[payload.HashPayload()] = payload
	logger.Printf("---- Received TX Payload ----: Height: %d -- Shard: %d \n", payload.Height, payload.ShardID)
	FileConnectionsLog.WriteString(fmt.Sprintf("---- Received TX Payload ----: Height: %d -- Shard: %d \n", payload.Height, payload.ShardID))
	payloadMap.Unlock()

	/*if(TxPayloadCointained(TransactionPayloadReceived,payload) == false){
		TransactionPayloadReceived = append(TransactionPayloadReceived,payload)
		processedTXPayloads = append(processedTXPayloads,payload.ShardID)
		//logger.Printf("Received Transaction Payload: %v\n", payload.StringPayload())
		logger.Printf("---- Received TX Payload ----: Height: %d -- Shard: %d \n", payload.Height, payload.ShardID)
		FileConnectionsLog.WriteString(fmt.Sprintf("---- Received TX Payload ----: Height: %d -- Shard: %d \n", payload.Height, payload.ShardID))
	}*/
}


func processEpochBlock(eb []byte) {
	var epochBlock *protocol.EpochBlock
	epochBlock = epochBlock.Decode(eb)
	logger.Printf("Received Epoch Block: %v\n", epochBlock.String())
	FileConnectionsLog.WriteString(fmt.Sprintf("Received Epoch Block: %v\n", epochBlock.String()))
	ValidatorShardMap = epochBlock.ValMapping
	NumberOfShards = epochBlock.NofShards
	ThisShardID = ValidatorShardMap.ValMapping[validatorAccAddress]
	lastEpochBlock = epochBlock
	//broadcastEpochBlock(lastEpochBlock)
	//p2p.EpochBlockOut <- eb
}

/*func processTXPayload(txPayload *protocol.TransactionPayload) (err error) {
	err = updateGlobalState(txPayload) //This function mimics the txPayload to update the global state in each shard, no Tx validation or storing done, because its done in the other shards
	return err
}*/

func processBlock(payload []byte) {
	var block *protocol.Block
	block = block.Decode(payload)

	if block.ShardId == ThisShardID && block.Height > lastEpochBlock.Height {
		//Block already confirmed and validated
		if storage.ReadClosedBlock(block.Hash) != nil {
			logger.Printf("Received block (%x) has already been validated.\n", block.Hash[0:8])
			FileConnectionsLog.WriteString(fmt.Sprintf("Received block (%x) has already been validated.\n", block.Hash[0:8]))
			return
		}

		//Start validation process
		err := validate(block, false)
		if err == nil {
			logger.Printf("Received Validated block: %vState:\n%v\n", block, getState())
			FileConnectionsLog.WriteString(fmt.Sprintf("Received Validated block: %vState:\n%v\n", block, getState()))
			broadcastBlock(block)
		} else {
			logger.Printf("Received block (%x) could not be validated: %v\n", block.Hash[0:8], err)
			FileConnectionsLog.WriteString(fmt.Sprintf("Received block (%x) could not be validated: %v\n", block.Hash[0:8], err))
		}
	} else {
		//broadcastBlock(block)
		storage.ReceivedBlockStash.Set(block.Hash,block)
		logger.Printf("Written block to stash Shard ID: %v  VS my shard ID: %v - Height: %d\n",block.ShardId,ThisShardID,block.Height)
		FileConnectionsLog.WriteString(fmt.Sprintf("Written block to stash Shard ID: %v  VS my shard ID: %v - Height: %d\n",block.ShardId,ThisShardID,block.Height))
		//if(lastBlock != nil){
		//	if(block.Height >= lastBlock.Height){
		//		storage.ReceivedBlockStash.Set(block.Hash,block)
		//	}
		//} else if(lastEpochBlock != nil) {
		//	if(block.Height >= lastEpochBlock.Height){
		//		storage.ReceivedBlockStash.Set(block.Hash,block)
		//	}
		//}
		//save hash of block for later creating epoch block. Make sure to store hashes from blocks other than my shard
		/*if(lastEpochBlock != nil){
			if(int(block.Height) == int(lastEpochBlock.Height) + activeParameters.epoch_length){
				lastShardMutex.Lock()
				LastShardHashesMap[block.Hash] = block.Hash
				lastShardMutex.Unlock()
				logger.Printf("Received lastshard hash for block height: %d\n",block.Height)
				FileConnectionsLog.WriteString(fmt.Sprintf("Received lastshard hash for block height: %d\n",block.Height))
			}
		}*/
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

/*Broadcast TX hashes to the network*/
func broadcastTxPayload() {
	p2p.TxPayloadOut <- TransactionPayloadOut.EncodePayload()
}

func broadcastEpochBlock(epochBlock *protocol.EpochBlock) {
	p2p.EpochBlockOut <- epochBlock.Encode()
}

//p2p.ValidatorShardMapChanOut is a channel whose data get consumed by the p2p package
func broadcastValidatorShardMapping(mapping *protocol.ValShardMapping) {
	p2p.ValidatorShardMapChanOut <- mapping.Encode()
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