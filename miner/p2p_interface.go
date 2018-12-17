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
		broadcastBlock(block)
	} else {
		logger.Printf("Received block (%x) could not be validated: %v\n", block.Hash[0:8], err)
	}
}

//p2p.BlockOut is a channel whose data get consumed by the p2p package
func broadcastBlock(block *protocol.Block) {
	p2p.BlockOut <- block.Encode()
	p2p.BlockHeaderOut <- block.EncodeHeader()
}

func broadcastVerifiedTxs(txs []*protocol.FundsTx) {
	var verifiedTxs [][]byte

	for _, tx := range txs {
		verifiedTxs = append(verifiedTxs, tx.Encode()[:])
	}

	p2p.VerifiedTxsOut <- protocol.Encode(verifiedTxs, protocol.FUNDSTX_SIZE)
}