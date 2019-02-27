package miner

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"time"
)

/**
	Helper function for logging purposes. Converts a slice of hashes to a human readable form
 */
func HashSliceToString(inputSlice [][32]byte) string {

	returnString := ""

	for _, hash := range inputSlice {
		returnString += fmt.Sprintf("(%x)",hash[0:8])
		returnString += " | "
	}

	return returnString
}
/**
	Helper function for logging purposes. Similarly to the above function, of commitment proofs to a human readable form
 */
func CommitmentProofSliceToString(inputSlice [][256]byte) string {

	returnString := ""

	for _, hash := range inputSlice {
		returnString += fmt.Sprintf("(%x)",hash[0:8])
		returnString += " | "
	}

	return returnString
}

func BlockSequenceToString(blockSeq []*protocol.Block) string {
	hashSlice := [][32]byte{}

	for _, block := range blockSeq {
		hashSlice = append(hashSlice,block.Hash)
	}

	return HashSliceToString(hashSlice)
}

//Function to give a list of blocks to rollback (in the right order) and a list of blocks to validate.
//Covers both cases (if block belongs to the longest chain or not).
func getBlockSequences(newBlock *protocol.Block) (blocksToRollback, blocksToValidate []*protocol.Block, err error) {
	//Fetch all blocks that are needed to validate.
	ancestorHash, newChain := getNewChain(newBlock)

	//Common ancestorHash not found, discard block.
	if ancestorHash == [32]byte{} {
		return nil, nil, errors.New("Common ancestorHash not found.")
	}

	//Count how many blocks there are on the currently active chain.
	tmpBlock := lastBlock

	for tmpBlock.Height > lastEpochBlock.Height{
		if tmpBlock.Hash == ancestorHash {
			break
		}
		blocksToRollback = append(blocksToRollback, tmpBlock)
		FileLogger.Printf("Added block (%x) to rollback blocks\n",tmpBlock.Hash[0:8])
		//The block needs to be in closed storage.
		tmpBlockNewHash := tmpBlock.PrevHash
		tmpBlock = storage.ReadClosedBlock(tmpBlockNewHash)
		if(tmpBlock != nil){
			FileLogger.Printf("New tmpBlock: (%x)\n",tmpBlock.Hash[0:8])
		} else {
			FileLogger.Printf("tmpBlock is nil. No Block found in closed storage for hash: (%x)\n",tmpBlockNewHash[0:8])
			if(ancestorHash == storage.ReadLastClosedEpochBlock().Hash){
				break
			}
		}
	}


	//Compare current length with new chain length.
	if len(blocksToRollback) >= len(newChain) {
		//Current chain length is longer or equal (our consensus protocol states that in this case we reject the block).
		return nil, nil, errors.New(fmt.Sprintf("Block belongs to shorter or equally long chain (blocks to rollback %d vs block of new chain %d)", len(blocksToRollback), len(newChain)))
	} else {
		//New chain is longer, rollback and validate new chain.
		return blocksToRollback, newChain, nil
	}
}


func getNewChain(newBlock *protocol.Block) (ancestor [32]byte, newChain []*protocol.Block) {
OUTER:
	for {
		newChain = append(newChain, newBlock)

		//Search for an ancestor (which needs to be in closed storage -> validated block).
		prevBlockHash := newBlock.PrevHash

		//Search in closed (Validated) blocks first
		potentialAncestor := storage.ReadClosedBlock(prevBlockHash)
		if potentialAncestor != nil {
			//Found ancestor because it is found in our closed block storage.
			//We went back in time, so reverse order.
			newChain = InvertBlockArray(newChain)
			return potentialAncestor.Hash, newChain
		} else {
			//Check if ancestor is an epoch block
			potentialEpochAncestorHash := storage.ReadLastClosedEpochBlock().Hash
			if prevBlockHash == potentialEpochAncestorHash {
				//Found ancestor because it is found in our closed block storage.
				//We went back in time, so reverse order.
				newChain = InvertBlockArray(newChain)
				return potentialEpochAncestorHash, newChain
			}
		}


		//It might be the case that we already started a sync and the block is in the openblock storage.
		newBlock = storage.ReadOpenBlock(prevBlockHash)
		if newBlock != nil {
			continue
		}

		// Check if block is in received stash. When in there, continue outer for-loop (Sorry for GO-TO), until ancestor
		// is found in closed block storage. The blocks from the stash will be validated in the normal validation process
		// after the rollback. (Similar like when in open storage) If not in stash, continue with a block request to
		// the network. Keep block in stash in case of multiple rollbacks (Very rare)
		for _, block := range storage.ReadReceivedBlockStash() {
			if block.Hash == prevBlockHash {
				newBlock = block
				continue OUTER
			}
		}

		//TODO Optimize code (duplicated)
		//Fetch the block we apparently missed from the network.
		p2p.BlockReq(prevBlockHash)

		//Blocking wait
		select {
		case encodedBlock := <-p2p.BlockReqChan:
			newBlock = newBlock.Decode(encodedBlock)
			storage.WriteToReceivedStash(newBlock)
			//Limit waiting time to BLOCKFETCH_TIMEOUT seconds before aborting.
		case <-time.After(BLOCKFETCH_TIMEOUT * time.Second):
			return [32]byte{}, nil
		}
	}

	return [32]byte{}, nil
}
