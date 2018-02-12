package miner

import (
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/sha3"
	"time"
)

//Tests whether the first diff bits are zero
func validateProofOfWork(diff uint8, hash [32]byte) bool {
	var byteNr uint8
	//Bytes check
	for byteNr = 0; byteNr < (uint8)(diff/8); byteNr++ {
		if hash[byteNr] != 0 {
			return false
		}
	}
	//Bits check
	if diff%8 != 0 && hash[byteNr+1] >= 1<<(8-diff%8) {
		return false
	}
	return true
}

//diff and partialHash is needed to calculate a valid PoW, prevHash is needed to check whether we should stop
//PoW calculation because another block has been validated meanwhile
func proofOfWork(diff uint8, partialHash, prevHash [32]byte) ([8]byte, error) {

	var (
		pow    [32]byte
		byteNr uint8
		abort  bool

		cntBuf [8]byte
		cnt    uint64
	)

	//2^64-1 = 18446744073709551615
	for cnt = 0; cnt < 18446744073709551615; cnt++ {
		//lastBlock is a global variable which points to the last block. This check makes sure we abort if another
		//block has been validated

		time.Sleep(time.Millisecond)

		if prevHash != lastBlock.Hash {
			return [8]byte{}, errors.New("Abort mining, another block has been successfully validated in the meantime")
		}
		abort = false

		binary.BigEndian.PutUint64(cntBuf[:], cnt)
		pow = sha3.Sum256(append(cntBuf[:], partialHash[:]...))
		//Byte check
		for byteNr = 0; byteNr < (uint8)(diff/8); byteNr++ {
			if pow[byteNr] != 0 {
				abort = true
				break
			}
		}
		if abort {
			continue
		}
		//Bit check
		if diff%8 != 0 && pow[byteNr+1] >= 1<<(8-diff%8) {
			continue
		}
		break
	}

	return cntBuf, nil
}
