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
	countConsolidationDistance := 0
	for {
		if tmpBlock.Hash == ancestor.Hash {
			break
		}

		fmt.Printf("the before height block %v %x needs to be in closed storage: tmpBlock:\n%v\n", tmpBlock.Height, tmpBlock.PrevHash, tmpBlock)
		//the block needs to be in closed storage
		if tmpBlock.NrConsolidationTx > 0 {
			txHash := tmpBlock.ConsolidationTxData[0]
			tx := getTransaction(p2p.CONSOLIDATIONTX_REQ, txHash)
			consolidationTx := tx.(*protocol.ConsolidationTx)
			fmt.Printf("Previous cons Hash: %x\n", consolidationTx)

			prevConsBlock := storage.ReadClosedBlock(consolidationTx.PreviousConsHash)
			countConsolidationDistance += int(tmpBlock.Height - prevConsBlock.Height)

			tmpBlock = prevConsBlock
			continue
		} else {
			blocksToRollback = append(blocksToRollback, tmpBlock)
		}
		//if tmpBlock.Hash == [32]byte{} && len(newChain) > 0 {
		//	break
		//}
		tmpBlock2 := storage.ReadClosedBlock(tmpBlock.PrevHash)
		if tmpBlock2 != nil {
			fmt.Printf("storage.ReadClosedBlock(tmpBlock.PrevHash):\n%v\n", tmpBlock2)
			tmpBlock = tmpBlock2
		}
	}
	fmt.Printf("len(blocksToRollback %v) >= (len(newChain %v) + countConsolidationDistance %v)\n", len(blocksToRollback) ,len(newChain) , countConsolidationDistance )
	//Compare current length with new chain length
	if len(blocksToRollback) >= (len(newChain) + countConsolidationDistance) {
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
	var prevBlockHash [32]byte
	newBlockHash := newBlock.Hash
	for {
		fmt.Printf("Processing newBlock:\n%v\n", newBlock)
		newChain = append(newChain, newBlock)
		//Search for an ancestor (which needs to be in closed storage -> validated block)
		if newBlock.Hash != newBlockHash && newBlock.NrConsolidationTx > 0 {
			txHash := newBlock.ConsolidationTxData[0]
			tx := getTransaction(p2p.CONSOLIDATIONTX_REQ, txHash)
			consolidationTx := tx.(*protocol.ConsolidationTx)
			prevBlockHash = consolidationTx.PreviousConsHash
		} else {
			prevBlockHash = newBlock.PrevHash
		}
		potentialAncestor := storage.ReadClosedBlock(prevBlockHash)
		if potentialAncestor != nil {
			fmt.Printf("found ancestor because it is found in our closed block storage\n")
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
		newBlock = getBlockSync(prevBlockHash)
		if newBlock == nil {
			return nil, nil
		}
	}
	return nil, nil
}


func getTransaction(txType uint8, txHash [32]byte) (consolidationTx protocol.Transaction){
	fmt.Printf("Getting tx type %v with hash %x\n", txType, txHash)
	if closedTx := storage.ReadClosedTx(txHash); closedTx != nil {
		return closedTx
	}
	tx := reqTx(txType, txHash)
	if tx == nil {
		fmt.Printf("Could not find transaction %x\n", txHash)
	}
	return tx
}

func getBlockSync(blockHash [32]byte) (newBlock *protocol.Block) {
	fmt.Printf("fetching block %v\n", blockHash)

	newBlock = storage.ReadClosedBlock(blockHash)
	if newBlock != nil {
		return newBlock
	}
	//Fetch the block we apparently missed from the network
	p2p.BlockReq(blockHash)

	//Blocking wait
	select {
	case encodedBlock := <-p2p.BlockReqChan:
		newBlock = newBlock.Decode(encodedBlock)
		return newBlock
		//Limit waiting time to BLOCKFETCH_TIMEOUT seconds before aborting
	case <-time.After(BLOCKFETCH_TIMEOUT * time.Second):
		fmt.Printf("Timeout while fetching block %x\n", blockHash)
		return nil
	}
}