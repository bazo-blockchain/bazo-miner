package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
)

type EpochBlock struct {
	//Header
	Header       		byte
	Hash         		[32]byte
	PrevShardHashes     [][32]byte
	Height       		uint32

	//Body
	Timestamp             int64
	MerkleRoot            [32]byte
	MerklePatriciaRoot    [32]byte
	CommitmentProof       [crypto.COMM_PROOF_LENGTH]byte
}

func NewEpochBlock(prevShardHashes [][32]byte, height uint32) *EpochBlock {
	newEpochBlock := EpochBlock{
		PrevShardHashes: prevShardHashes,
		Height:   height,
	}

	return &newEpochBlock
}

func (epochBlock *EpochBlock) HashEpochBlock() [32]byte {
	if epochBlock == nil {
		return [32]byte{}
	}

	blockHash := struct {
		prevShardHashes               [][32]byte
		timestamp             		  int64
		merkleRoot            		  [32]byte
		merklePatriciaRoot	  		  [32]byte
		height				  		  uint32
		commitmentProof       [crypto.COMM_PROOF_LENGTH]byte
	}{
		epochBlock.PrevShardHashes,
		epochBlock.Timestamp,
		epochBlock.MerkleRoot,
		epochBlock.MerklePatriciaRoot,
		epochBlock.Height,
		epochBlock.CommitmentProof,
	}
	return SerializeHashContent(blockHash)
}

func (epochBlock *EpochBlock) Encode() []byte {
	if epochBlock == nil {
		return nil
	}

	encoded := EpochBlock{
		Header:                epochBlock.Header,
		Hash:                  epochBlock.Hash,
		PrevShardHashes:       epochBlock.PrevShardHashes,
		Timestamp:             epochBlock.Timestamp,
		MerkleRoot:            epochBlock.MerkleRoot,
		MerklePatriciaRoot:    epochBlock.MerklePatriciaRoot,
		Height:                epochBlock.Height,
		CommitmentProof:	   epochBlock.CommitmentProof,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (epochBlock *EpochBlock) EncodeHeader() []byte {
	if epochBlock == nil {
		return nil
	}

	encoded := EpochBlock{
		Header:       		 epochBlock.Header,
		Hash:         		 epochBlock.Hash,
		PrevShardHashes:     epochBlock.PrevShardHashes,
		Height:       		 epochBlock.Height,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (epochBlock *EpochBlock) Decode(encoded []byte) (b *EpochBlock) {
	if encoded == nil {
		return nil
	}

	var decoded EpochBlock
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (epochBlock EpochBlock) String() string {
	return fmt.Sprintf("\nHash: %x\n"+
		"Len Previous Shard Hashes: %d\n"+
		"Prev Shard Hashes: %x\n"+
		"Timestamp: %v\n"+
		"MerkleRoot: %x\n"+
		"MerklePatriciaRoot: %x\n"+
		"Height: %d\n"+
		"Commitment Proof: %x\n",
		epochBlock.Hash[0:8],
		len(epochBlock.PrevShardHashes),
		epochBlock.StringPrevHashes(),
		epochBlock.Timestamp,
		epochBlock.MerkleRoot[0:8],
		epochBlock.MerklePatriciaRoot,
		epochBlock.Height,
		epochBlock.CommitmentProof[0:8],
	)
}

func (epochBlock EpochBlock) StringPrevHashes() (prevHashes string) {

	for _, prevHash := range epochBlock.PrevShardHashes {
		prevHashes += fmt.Sprintf("%x\n", prevHash)
	}
	return prevHashes
}
