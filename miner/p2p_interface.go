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

		//check validator assignment
		validatorMapping := <- p2p.ValidatorShardMapChanIn
		processReceivedValidatorMapping(validatorMapping)

		//receive Epoch Block
		epochBlock := <-p2p.EpochBlockIn
		processEpochBlock(epochBlock)

		//TX payload in
		txPayload := <- p2p.TxPayloadIn
		processPayload(txPayload)
	}
}
func processReceivedValidatorMapping(vm []byte) {
	var valMapping *protocol.ValShardMapping
	valMapping = valMapping.Decode(vm)
	ValidatorShardMap = valMapping
	broadcastValidatorShardMapping(valMapping)
}
func processPayload(payloadByte []byte) {
	var payload *protocol.TransactionPayload
	payload = payload.DecodePayload(payloadByte)

	/*save TxPayloads from other shards and only process TxPayloads from shards which haven't been processed yet*/

	if(blockBeingProcessed != nil){
		if payload.ShardID != ThisShardID && IntSliceContains(processedTXPayloads, payload.ShardID) == false && +
			payload.Height == int(blockBeingProcessed.Height) {
			TransactionPayloadIn = append(TransactionPayloadIn,payload)
			processedTXPayloads = append(processedTXPayloads,payload.ShardID)
			logger.Printf("Received Transaction Payload: %v\n", payload.StringPayload())
		}
	}
	/*if (IntSliceContains(processedTXPayloads, payload.ShardID) == false){
		err := processTXPayload(payload)
		if err != nil{
			logger.Printf("error while processing transaction payload: %v\n",err)
		} else {
			processedTXPayloads = append(processedTXPayloads,payload.ShardID)
		}
	}*/
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

	if block.ShardId == ThisShardID {
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
		/*if(lastBlock == nil && lastEpochBlock != nil){
			if(block.Height == lastEpochBlock.Height + 1 && HashSliceContains(LastShardHashes,block.Hash) == false){
				//count blocks received at current height
				ReceivedBlocksAtHeightX = ReceivedBlocksAtHeightX + 1

				//save hash of block for later creating epoch block. Make sure to store hashes from blocks other than my shard
				LastShardHashes = append(LastShardHashes, block.Hash)
			}
		} else */
		if lastBlock != nil {
			if block.Height == lastBlock.Height + 1 && HashSliceContains(LastShardHashes,block.Hash) == false {
				//count blocks received at current height
				ReceivedBlocksAtHeightX = ReceivedBlocksAtHeightX + 1

				//save hash of block for later creating epoch block. Make sure to store hashes from blocks other than my shard
				LastShardHashes = append(LastShardHashes, block.Hash)
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