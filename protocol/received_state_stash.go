package protocol

import (
	"sync"
)
/*This datastructe maintains a map of the form [32]byte - *StateTransition. It stores the state transitions received from other shards.
This datastructure will be queried at every blockheight to check if we can continue mining the next block.
Because we need to remove the first element of this datastructure and map access is random in Go, we additionally have a slice datastructure
which keeps track of the order of the included state transition. Such that, using the slice structure, we can remove the first received block once this
stash gets full*/
type KeyState [32]byte   // Key: Hash of the block
type ValueState *StateTransition // Value: Block

type StateStash struct {
	M    map[KeyState]ValueState
	Keys []KeyState
}

var stateMutex			= &sync.Mutex{}

func NewStateStash() *StateStash {
	return &StateStash{M: make(map[KeyState]ValueState)}
}

/*This function includes a key and tracks its order in the slice*/
func (m *StateStash) Set(k KeyState, v ValueState) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	/*Check if the map does not contain the key*/
	if _, ok := m.M[k]; !ok {
		m.Keys = append(m.Keys, k)
		m.M[k] = v
	}

	/*When lenght of stash is > 50 --> Remove first added Block*/
	if(len(m.M) > 50){
		m.DeleteFirstEntry()
	}
}

/*This function includes a key and tracks its order in the slice. No need to put the lock because it is used from the calling function*/
func (m *StateStash) DeleteFirstEntry() {
	/*stashMutex.Lock()
	defer stashMutex.Unlock()*/
	firstStateTransitionHash := m.Keys[0]

	if _, ok := m.M[firstStateTransitionHash]; ok {
		delete(m.M,firstStateTransitionHash)
	}
	m.Keys = append(m.Keys[:0], m.Keys[1:]...)
}

/*This function counts how many state transisitions in the stash have some predefined height*/
func CheckForHeightStateTransition(statestash *StateStash, height uint32) int {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	numberOfStateTransisionsAtHeight := 0
	for _,stateTransision := range statestash.M {
		if(stateTransision.Height == int(height)){
			numberOfStateTransisionsAtHeight = numberOfStateTransisionsAtHeight + 1
		}
	}
	return numberOfStateTransisionsAtHeight
}

func ReturnStateTransitionForHeight(statestash *StateStash, height uint32) [] *StateTransition {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	stateTransitionSlice := []*StateTransition{}

	for _,st := range statestash.M {
		if(st.Height == int(height)){
			stateTransitionSlice = append(stateTransitionSlice,st)
		}
	}

	return stateTransitionSlice
}

func ReturnShardHashesForHeight(statestash *StateStash, height uint32) [][32]byte {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	hashSlice := [][32]byte{}

	for _,st := range statestash.M {
		if(st.Height == int(height)){
			hashSlice = append(hashSlice,st.BlockHash)
		}
	}

	return hashSlice
}

func ReturnStateTransitionForPosition(stateStash *StateStash, position int) (stateHash [32]byte, stateTransition *StateTransition) {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	if(position > len(stateStash.Keys)-1){
		return [32]byte{}, nil
	}

	stateStashPos := stateStash.Keys[position]

	return stateStashPos, stateStash.M[stateStashPos]
}

///*This function returns the hashes of the blocks for some height*/
//func ReturnHashesForHeight(blockstash *BlockStash, height uint32) (hashes [][32]byte) {
//	stashMutex.Lock()
//	defer stashMutex.Unlock()
//	var blockHashes [][32]byte
//
//	for _,block := range blockstash.M {
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
//	for _,block := range blockstash.M {
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
//	if(position > len(blockstash.Keys)-1){
//		return [32]byte{}, nil
//	}
//
//	blockHashPos := blockstash.Keys[position]
//
//	return blockHashPos, blockstash.M[blockHashPos]
//}