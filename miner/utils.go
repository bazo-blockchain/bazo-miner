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
func InvertEpochBlockArray(array []*protocol.EpochBlock) []*protocol.EpochBlock {
	for i, j := 0, len(array)-1; i < j; i, j = i+1, j-1 {
		array[i], array[j] = array[j], array[i]
	}

	return array
}