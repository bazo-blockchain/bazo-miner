package miner

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"time"
)

//Function to give a list of blocks to rollback (in the right order) and a list of blocks to validate.
//Covers both cases (if block belongs to the longest chain or not).
func getBlockSequences(newBlock *protocol.Block) (blocksToRollback, blocksToValidate []*protocol.Block, err error) {
	//Fetch all blocks that are needed to validate.
	ancestor, newChain := getNewChain(newBlock)
	//Common ancestor not found, discard block.
	if ancestor == nil {
		return nil, nil, errors.New("Common ancestor not found.")
	}
	logger.Printf("BLOCK_SEQENCE NewBlock %x --> ancestor: %x", newBlock.Hash[0:8], ancestor.Hash[0:8])

	//Count how many blocks there are on the currently active chain.
	tmpBlock := lastBlock

	for {
		if tmpBlock.Hash == ancestor.Hash {
			break
		}
		blocksToRollback = append(blocksToRollback, tmpBlock)
		//The block needs to be in closed storage.
		tmpBlock = storage.ReadClosedBlock(tmpBlock.PrevHash)
	}

	//Compare current length with new chain length.
	if len(blocksToRollback) >= len(newChain) {
		//Current chain length is longer or equal (our consensus protocol states that in this case we reject the block).
		return nil, nil, errors.New(fmt.Sprintf("Block belongs to shorter or equally long chain --> NO ROLLBACK (blocks to rollback %d vs block of new chain %d)", len(blocksToRollback), len(newChain)))
	} else {
		//New chain is longer, rollback and validate new chain.
		logger.Printf("BLOCK_ROLLBACK %d Block's to roll back: [", len(blocksToRollback))
		for _, block := range blocksToRollback {
			//logger.Printf("Rolled back block: %vState:\n%v", block, getState())
			logger.Printf("%x", block.Hash[0:8])
		}
		logger.Printf("]")
		return blocksToRollback, newChain, nil
	}
}

//Returns the ancestor from which the split occurs (if a split occurred, if not it's just our last block) and a list
//of blocks that belong to a new chain.
func getNewChain(newBlock *protocol.Block) (ancestor *protocol.Block, newChain []*protocol.Block) {
	for {
		newChain = append(newChain, newBlock)

		//Search for an ancestor (which needs to be in closed storage -> validated block).
		prevBlockHash := newBlock.PrevHash
		logger.Printf("SEARCHING for Block: %x --> START", prevBlockHash[0:8])
		potentialAncestor := storage.ReadClosedBlock(prevBlockHash)

		if potentialAncestor != nil {
			//Found ancestor because it is found in our closed block storage.
			//We went back in time, so reverse order.
			newChain = InvertBlockArray(newChain)

			return potentialAncestor, newChain
		}

		logger.Printf("SEARCHING for Block: %x --> NOT in CLOSED_BLOCK", prevBlockHash[0:8])

		//It might be the case that we already started a sync and the block is in the openblock storage.
		newBlock = storage.ReadOpenBlock(prevBlockHash)
		if newBlock != nil {
			logger.Printf("SEARCHING for Block: %x --> FOUND in OPEN_BLOCK", prevBlockHash[0:8])
			continue
		}

		logger.Printf("SEARCHING for Block: %x --> NOT FOUND in OPEN_BLOCK", prevBlockHash[0:8])
		//TODO Optimize code (duplicated)
		//Fetch the block we apparently missed from the network.

		logger.Printf("SEARCHING for Block: %x --> START Network Request", prevBlockHash[0:8])
		p2p.BlockReq(prevBlockHash)

		//Blocking wait
		select {
		case encodedBlock := <-p2p.BlockReqChan:
			newBlock = newBlock.Decode(encodedBlock)
			logger.Printf("RECEIVED_REQ Block: %x for BLOCK_REQ: %x", newBlock.Hash[0:8], prevBlockHash[0:8])
			//Limit waiting time to BLOCKFETCH_TIMEOUT seconds before aborting.
		case <-time.After(BLOCKFETCH_TIMEOUT * time.Second):
			logger.Printf("BLOCKFETCH_TIMEOUT occured while fetching block (%x) from network -> Return nil nil (Common ancestor probably not found)", prevBlockHash[0:8])
			return nil, nil
		}
	}

	return nil, nil
}
