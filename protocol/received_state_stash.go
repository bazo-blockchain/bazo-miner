package protocol

import (
	"sync"
)
/*This datastructe maintains a map of the form [32]byte - *Block. It stores the block received from other shards.
This datastructure will be queried at every blockheight to check if we can continue mining the next block.
Because we need to remove the first element of this datastructure and map access is random in Go, we additionally have a slice datastructure
which keeps track of the order of the included blocks. Such that, using the slice structure, we can remove the first received block once this
stash gets full*/
type KeyState [32]byte   // Key: Hash of the block
type ValueState *StateTransition // Value: Block

type StateStash struct {
	m    map[KeyState]ValueState
	keys []KeyState
}

var stateMutex			= &sync.Mutex{}

func NewStateStash() *StateStash {
	return &StateStash{m: make(map[KeyState]ValueState)}
}

/*This function includes a key and tracks its order in the slice*/
func (m *StateStash) Set(k KeyState, v ValueState) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	/*Check if the map does not contain the key*/
	if _, ok := m.m[k]; !ok {
		m.keys = append(m.keys, k)
		m.m[k] = v
	}

	/*When lenght of stash is > 50 --> Remove first added Block*/
	if(len(m.m) > 50){
		m.DeleteFirstEntry()
	}
}

/*This function includes a key and tracks its order in the slice. No need to put the lock because it is used from the calling function*/
func (m *StateStash) DeleteFirstEntry() {
	/*stashMutex.Lock()
	defer stashMutex.Unlock()*/
	firstStateTransitionHash := m.keys[0]

	if _, ok := m.m[firstStateTransitionHash]; !ok {
		delete(m.m,firstStateTransitionHash)
	}
	m.keys = append(m.keys[:0], m.keys[1:]...)
}

/*This function counts how many state transisitions in the stash have some predefined height*/
func CheckForHeightStateTransition(statestash *StateStash, height uint32) int {
	stashMutex.Lock()
	defer stashMutex.Unlock()
	numberOfStateTransisionsAtHeight := 0
	for _,stateTransision := range statestash.m {
		if(stateTransision.Height == int(height)){
			numberOfStateTransisionsAtHeight = numberOfStateTransisionsAtHeight + 1
		}
	}
	return numberOfStateTransisionsAtHeight
}

///*This function returns the hashes of the blocks for some height*/
//func ReturnHashesForHeight(blockstash *BlockStash, height uint32) (hashes [][32]byte) {
//	stashMutex.Lock()
//	defer stashMutex.Unlock()
//	var blockHashes [][32]byte
//
//	for _,block := range blockstash.m {
//		if(block.Height == height){
//			blockHashes = append(blockHashes,block.Hash)
//		}
//	}
//	return blockHashes
//}
//
///*This function extracts the transaction hashes of the blocks for some height*/
//func ReturnTxPayloadForHeight(blockstash *BlockStash, height uint32) (txpayload []*TransactionPayload) {
//	stashMutex.Lock()
//	defer stashMutex.Unlock()
//	payloadSlice := []*TransactionPayload{}
//
//	for _,block := range blockstash.m {
//		if(block.Height == height){
//			payload := NewTransactionPayload(block.ShardId,int(block.Height),nil,nil,nil,nil)
//			payload.StakeTxData = block.StakeTxData
//			payload.ConfigTxData = block.ConfigTxData
//			payload.FundsTxData = block.FundsTxData
//			payload.ContractTxData = block.ContractTxData
//			payloadSlice = append(payloadSlice,payload)
//		}
//	}
//	return payloadSlice
//}
//
///*This function extracts the item at some position*/
//func ReturnItemForPosition(blockstash *BlockStash, position int) (blockHash [32]byte, block *Block) {
//	stashMutex.Lock()
//	defer stashMutex.Unlock()
//
//	if(position > len(blockstash.keys)-1){
//		return [32]byte{}, nil
//	}
//
//	blockHashPos := blockstash.keys[position]
//
//	return blockHashPos, blockstash.m[blockHashPos]
//}