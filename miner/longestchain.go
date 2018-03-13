package miner

import (
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"time"
	"errors"
	"fmt"
)

//Function to give a list of blocks to rollback (in the right order) and a list of blocks to validate.
//Covers both cases (if block belongs to the longest chain or not to the longest chain)
func getBlockSequences(newBlock *protocol.Block) (blocksToRollback, blocksToValidate []*protocol.Block, err error) {

	//Fetch all blocks that are needed to validate
	ancestor, newChain := getNewChain(newBlock)

	//Common ancestor not found, discard block
	if ancestor == nil {
		return nil, nil, errors.New("common ancestor not found")
	}

	//Count how many blocks there are on the currently active chain
	tmpBlock := lastBlock
	for {
		if tmpBlock.Hash == ancestor.Hash {
			break
		}
		blocksToRollback = append(blocksToRollback, tmpBlock)
		//the block needs to be in closed storage
		tmpBlock = storage.ReadClosedBlock(tmpBlock.PrevHash)
	}

	//Compare current length with new chain length
	if len(blocksToRollback) >= len(newChain) {
		//Current chain length is longer or equal (our consensus protocol states that in this case we reject the block)
		return nil, nil, errors.New(fmt.Sprintf("block belongs to shorter or equally long chain (blocks to rollback %d vs block of new chain %d)", len(blocksToRollback), len(newChain)))
	} else {
		//New chain is longer, rollback and validate new chain
		return blocksToRollback, newChain, nil
	}
}

//Returns the ancestor from which the split occurs (if a split occurred, if not it's just our last block) and a list
//of blocks that belong to a new chain
func getNewChain(newBlock *protocol.Block) (ancestor *protocol.Block, newChain []*protocol.Block) {

	for {
		newChain = append(newChain, newBlock)

		//Search for an ancestor (which needs to be in closed storage -> validated block)
		prevBlockHash := newBlock.PrevHash
		potentialAncestor := storage.ReadClosedBlock(prevBlockHash)

		if potentialAncestor != nil {
			//found ancestor because it is found in our closed block storage
			//we went back in time, so reverse order
			newChain = InvertBlockArray(newChain)

			return potentialAncestor, newChain
		}

		//It might be the case that we already started a sync and the block is in the openblock storage
		newBlock = storage.ReadOpenBlock(prevBlockHash)
		if newBlock != nil {
			continue
		}
		//Fetch the block we apparently missed from the network
		p2p.BlockReq(prevBlockHash)

		//Blocking wait
		select {
		case encodedBlock := <-p2p.BlockReqChan:
			newBlock = newBlock.Decode(encodedBlock)
			//Limit waiting time to BLOCKFETCH_TIMEOUT seconds before aborting
		case <-time.After(BLOCKFETCH_TIMEOUT * time.Second):
			return nil, nil
		}
	}
	return nil, nil
}
