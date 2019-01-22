package miner

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"golang.org/x/crypto/sha3"
	"time"
)

//Tests whether the first diff bits are zero
func validateProofOfStake(diff uint8, prevSeeds [][32]byte, height uint32, balance uint64, seed [32]byte, timestamp int64) bool {
	var hashArgs []byte
	var heightBuf [4]byte
	var timestampBuf [8]byte

	binary.BigEndian.PutUint32(heightBuf[:], height)
	binary.BigEndian.PutUint64(timestampBuf[:], uint64(timestamp))

	//allocate memory
	//n * 32 bytes (prevSeeds) + 32 bytes (localSeed)+ 4 bytes (height) + 8 bytes (count)
	hashArgs = make([]byte, len(prevSeeds)*32+32+4+8)

	index := 0
	for _, prevSeed := range prevSeeds {
		copy(hashArgs[index:index+32], prevSeed[:])
		index += 32
	}

	copy(hashArgs[index:index+32], seed[:])
	copy(hashArgs[index+32:index+36], heightBuf[:])
	copy(hashArgs[index+36:index+44], timestampBuf[:])

	//calculate the hash
	pos := sha3.Sum256(hashArgs[:])

	data := binary.BigEndian.Uint64(pos[:])
	data = data / balance
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, data)

	copy(pos[0:32], buf.Bytes())

	var byteNr uint8
	//Bytes check
	for byteNr = 0; byteNr < (uint8)(diff/8); byteNr++ {
		if pos[byteNr] != 0 {
			return false
		}
	}
	//Bits check
	if diff%8 != 0 && pos[byteNr] >= 1<<(8-diff%8) {
		return false
	}
	return true
}

//diff and partialHash is needed to calculate a valid PoW, prevHash is needed to check whether we should stop
//PoW calculation because another block has been validated meanwhile

func proofOfStake(diff uint8, prevHash [32]byte, prevSeeds [][32]byte, height uint32, balance uint64, localSeed [32]byte) (int64, error) {

	var (
		pos    [32]byte
		byteNr uint8
		abort  bool

		timestampBuf [8]byte
		heightBuf    [4]byte

		timestamp int64

		hashArgs []byte
	)

	//allocate memory
	//n * 32 bytes (prevSeeds) + 32 bytes (localSeed)+ 4 bytes (height) + 8 bytes (count)
	hashArgs = make([]byte, len(prevSeeds)*32+32+4+8)

	binary.BigEndian.PutUint32(heightBuf[:], height)

	//all required parameters are concatinated in the following order:
	//([PrevSeeds] ⋅ LocalSeed ⋅ CurrentBlockHeight ⋅ Seconds)
	index := 0
	for _, prevSeed := range prevSeeds {
		copy(hashArgs[index:index+32], prevSeed[:])
		index += 32
	}
	copy(hashArgs[index:index+32], localSeed[:])    //32 bytes
	copy(hashArgs[index+32:index+36], heightBuf[:]) //4 bytes

	for _ = range time.Tick(time.Second) {
		//lastBlock is a global variable which points to the last block. This check makes sure we abort if another
		//block has been validated
		if prevHash != lastBlock.Hash {
			return -1, errors.New("Abort mining, another block has been successfully validated in the meantime")
		}

		abort = false

		//add the number of seconds that have passed since the Unix epoch (00:00:00 UTC, 1 January 1970)
		timestamp = time.Now().Unix()
		binary.BigEndian.PutUint64(timestampBuf[:], uint64(timestamp))
		copy(hashArgs[index+36:index+44], timestampBuf[:]) //8 bytes

		//calculate the hash
		pos = sha3.Sum256(hashArgs[:])

		//divide the hash by the balance (should not happen but possible in a testing environment)
		data := binary.BigEndian.Uint64(pos[:])
		if balance == 0 {
			return -1, errors.New("Zero division: Account owns 0 coins.")
		}
		data = data / balance
		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, data)

		copy(pos[0:32], buf.Bytes())

		//TODO @simibac What do you do here?
		//Byte check
		for byteNr = 0; byteNr < (uint8)(diff/8); byteNr++ {
			if pos[byteNr] != 0 {
				//continue begins the next iteration of the innermost
				abort = true
				break
			}
		}

		if abort {
			continue
		}
		//Bit check
		if diff%8 != 0 && pos[byteNr] >= 1<<(8-diff%8) {
			continue
		}
		break
	}

	return timestamp, nil
}

func GetLatestSeeds(n int, block *protocol.Block) [][32]byte {
	var prevSeeds [][32]byte
	for block.Height > 0 && n > 0 {
		block = storage.ReadClosedBlock(block.PrevHash)
		prevSeeds = append(prevSeeds, block.Seed)
		n -= 1
	}
	return prevSeeds
}
