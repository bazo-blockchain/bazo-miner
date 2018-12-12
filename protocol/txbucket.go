package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type TxBucket struct {
	Hash		[32]byte
	Address		[64]byte
	Amount		uint64
	MerkleRoot	[32]byte
	FundsTxData [][32]byte
}

func NewTxBucket(address [64]byte) *TxBucket {
	return &TxBucket{
		Address:	address,
	}
}

func (bucket *TxBucket) CreateHash() [32]byte {
	if bucket == nil {
		return [32]byte{}
	}

	bucketHash := struct {
		hash       [32]byte
		address    [64]byte
		amount		uint64
		merkleRoot [32]byte
	}{
		bucket.Hash,
		bucket.Address,
		bucket.Amount,
		bucket.MerkleRoot,
	}

	return SerializeHashContent(bucketHash)
}

func (bucket *TxBucket) Encode() []byte {
	if bucket == nil {
		return nil
	}

	encoded := TxBucket{
		Hash:			bucket.Hash,
		Address:		bucket.Address,
		MerkleRoot:     bucket.MerkleRoot,
		Amount:			bucket.Amount,
		FundsTxData:	bucket.FundsTxData,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (bucket *TxBucket) Decode(encoded []byte) (b *TxBucket) {
	if encoded == nil {
		return nil
	}

	var decoded TxBucket
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (bucket TxBucket) String() string {
	return fmt.Sprintf("\nHash: %x\n" +
		"Address: %x\n" +
		"Amount: %v\n"+
		"MerkleRoot: %x\n",
		bucket.Hash[0:8],
		bucket.Address[0:8],
		bucket.Amount,
		bucket.MerkleRoot[0:8],
	)
}