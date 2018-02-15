package miner

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

func InvertBlockArray(array []*protocol.Block) []*protocol.Block {
	for i, j := 0, len(array)-1; i < j; i, j = i+1, j-1 {
		array[i], array[j] = array[j], array[i]
	}

	return array
}

func Prepend(arr [][32]byte, item [32]byte) [][32]byte {
	return append([][32]byte{item}, arr...)
}
