package miner

import (
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
	//broadcastValidatorShardMapping(valMapping)
}
func processPayload(payloadByte []byte) {
	var payload *protocol.TransactionPayload
	payload = payload.DecodePayload(payloadByte)


	payloadMap.Lock()
	TransactionPayloadReceivedMap[payload.HashPayload()] = payload
	payloadMap.Unlock()

	if(TxPayloadCointained(TransactionPayloadReceived,payload) == false){
		TransactionPayloadReceived = append(TransactionPayloadReceived,payload)
		processedTXPayloads = append(processedTXPayloads,payload.ShardID)
		logger.Printf("Received Transaction Payload: %v\n", payload.StringPayload())
	}
}


func processEpochBlock(eb []byte) {
	var epochBlock *protocol.EpochBlock
	epochBlock = epochBlock.Decode(eb)
	logger.Printf("Received Epoch Block: %v\n", epochBlock.String())
	lastEpochBlock = epochBlock
}

/*func processTXPayload(txPayload *protocol.TransactionPayload) (err error) {
	err = updateGlobalState(txPayload) //This function mimics the txPayload to update the global state in each shard, no Tx validation or storing done, because its done in the other shards
	return err
}*/

func processBlock(payload []byte) {
	var block *protocol.Block
	block = block.Decode(payload)

	logger.Printf("Incoming block hehe:")
	logger.Printf("Incoming block Shard ID: %v",block.ShardId)
	logger.Printf("VS my shard ID: %d",ThisShardID)
	logger.Printf("Incoming block Height: %v",block.Height)

	if block.ShardId == ThisShardID && block.Height > lastEpochBlock.Height {
		//Block already confirmed and validated
		if storage.ReadClosedBlock(block.Hash) != nil {
			logger.Printf("Received block (%x) has already been validated.\n", block.Hash[0:8])
			return
		}

		//Start validation process
		err := validate(block, false)
		if err == nil {
			logger.Printf("Validated block: %vState:\n%v", block, getState())
			broadcastBlock(block)
		} else {
			logger.Printf("Received block (%x) could not be validated: %v\n", block.Hash[0:8], err)
		}
	} else {
		//save hash of block for later creating epoch block. Make sure to store hashes from blocks other than my shard
		if(lastEpochBlock != nil){
			if(int(block.Height) == int(lastEpochBlock.Height) + activeParameters.epoch_length && HashSliceContains(LastShardHashes,block.Hash) == false){
				LastShardHashes = append(LastShardHashes, block.Hash)
				logger.Printf("Received lastshard hash for block height: %d",block.Height)
			}
		}
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