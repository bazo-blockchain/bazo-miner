package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"sync"
)

type State struct {
	//Header
	ActualState       map[[64]byte]*Account
	StateLock		  sync.Mutex
}

func NewState() *State {
	newState := new(State)
	newState.ActualState = make(map[[64]byte]*Account)
	newState.StateLock = sync.Mutex{}
	return newState
}

func (state *State) HashState() [32]byte {
	if state == nil {
		return [32]byte{}
	}

	stateHash := struct {
		state				  map[[64]byte]*Account
	}{
		state.ActualState,
	}
	return SerializeHashContent(stateHash)
}

func (state *State) GetStateSize() uint64 {
	size := len(state.ActualState)
	return uint64(size)
}

func (state *State) Encode() []byte {
	if state == nil {
		return nil
	}

	encoded := State{
		ActualState:                state.ActualState,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (state *State) Decode(encoded []byte) (s *State) {
	if encoded == nil {
		return nil
	}

	var decoded State
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (state State) String() string {
	stateString := ""
	for _, acc := range state.ActualState {
		stateString += fmt.Sprintf("Is root: %v, %v\n", storage.IsRootKey(acc.Address), acc)
	}
	return stateString
}
