package miner

import (
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

var (
	receivedBlockStash = make([]*protocol.Block, 0)
)
//The code in this source file communicates with the p2p package via channels

//Constantly listen to incoming data from the network
func incomingData() {
	for {
		block := <-p2p.BlockIn
		//logger.Printf("New Incoming Block")
		processBlock(block)
	}
}

func blockAlreadyInStash(slice []*protocol.Block, newBlockHash [32]byte) bool {
	for _, blockInStash := range slice {
		if blockInStash.Hash == newBlockHash {
			return true
		}
	}
	return false
}

//ReceivedBlockStash is a stash with all Blocks received such that we can prevent forking
func processBlock(payload []byte) {
	var block *protocol.Block
	block = block.Decode(payload)

	logger.Printf("RECEIVED block: %x ", block.Hash[0:8])
	//Block already confirmed and validated
	if storage.ReadClosedBlock(block.Hash) != nil {
		logger.Printf("Received block (%x) has already been validated.\n", block.Hash[0:8])
		return
	}

	//Append received Block to stash, when it is not there already & keep size at 20
	if !blockAlreadyInStash(receivedBlockStash, block.Hash) {
		receivedBlockStash = append(receivedBlockStash, block)
		if len(receivedBlockStash) > 40 {
			receivedBlockStash = append(receivedBlockStash[:0], receivedBlockStash[1:]...)
		}
	}

	//Print stash
	logger.Printf("RECEIVED_BLOCK_STASH: Length: %v, [", len(receivedBlockStash))
	for _, block := range receivedBlockStash {
		logger.Printf("%x", block.Hash[0:8])
	}
	logger.Printf("]")

	//Start validation process
	err := validate(block, false)
	if err == nil {
		logger.Printf("Validated block: %vState:\n%v", block, getState())
		logger.Printf("BROADCAST received block: %x ", block.Hash[0:8])
		broadcastBlock(block)
		CalculateBlockchainSize(block.GetSize())
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

func broadcastVerifiedTxs(txs []*protocol.FundsTx) {
	var verifiedTxs [][]byte

	for _, tx := range txs {
		verifiedTxs = append(verifiedTxs, tx.Encode()[:])
	}

	p2p.VerifiedTxsOut <- protocol.Encode(verifiedTxs, protocol.FUNDSTX_SIZE)
}