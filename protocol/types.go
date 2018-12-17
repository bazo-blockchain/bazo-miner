package protocol

import (
	"bytes"
	"log"
)

type ByteArray 		[]byte
type AddressType 	[64]byte
type HashType 		[32]byte
type HashArray 		[]HashType

func (h HashArray) Len() int {
	return len(h)
}

func (h HashArray) Less(i, j int) bool {
	switch bytes.Compare(h[i][:], h[j][:]) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		log.Panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
		return false
	}
}

func (h HashArray) Swap(i, j int) {
	h[j], h[i] = h[i], h[j]
}


