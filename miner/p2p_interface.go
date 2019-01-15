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
		var valMapping *protocol.ValShardMapping
		valMapping = valMapping.Decode(validatorMapping)
		broadcastValidatorShardMapping(valMapping)
	}
}

func processBlock(payload []byte) {
	var block *protocol.Block
	block = block.Decode(payload)

	//Block already confirmed and validated
	if storage.ReadClosedBlock(block.Hash) != nil {
		logger.Printf("Received block (%x) has already been validated.\n", block.Hash[0:8])
		return
	}

	//Start validation process
	err := validate(block, false)
	if err == nil {
		logger.Printf("Validated block: %vState:\n%v", block, getState())

		if(block.Height == lastBlock.Height && block.ShardId != ThisShardID && HashSliceContains(LastShardHashes,block.Hash) == true){
			//count blocks received at current height
			ReceivedBlocksAtHeightX = ReceivedBlocksAtHeightX + 1

			//save hash of block for later creating epoch block. Make sure to store hashes from blocks other than my shard
			LastShardHashes = append(LastShardHashes, block.Hash)
		}

		broadcastBlock(block)
	} else {
		logger.Printf("Received block (%x) could not be validated: %v\n", block.Hash[0:8], err)
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