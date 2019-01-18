package protocol

import (
	"bytes"
	"encoding/gob"
)

type TxBucket struct {
	Address         [64]byte
	RelativeAmount int64

	merkleRoot HashType

	Transactions []*FundsTx // won't be serialized
}

func NewTxBucket(address [64]byte) *TxBucket {
	return &TxBucket{
		Address:	address,
	}
}

func (bucket *TxBucket) AddFundsTx(tx *FundsTx) {
	if tx.From == bucket.Address {
		bucket.RelativeAmount += int64(tx.Amount + tx.Fee)
	} else {
		return
	}

	bucket.Transactions = append(bucket.Transactions, tx)
}

func (bucket *TxBucket) Hash() HashType {
	if bucket == nil {
		return HashType{}
	}

	bucketHash := struct {
		address    [64]byte
		amount     int64
	}{
		bucket.Address,
		bucket.RelativeAmount,
	}

	return SerializeHashContent(bucketHash)
}

func (bucket *TxBucket) Encode() []byte {
	if bucket == nil {
		return nil
	}

	encoded := TxBucket{
		Address:         bucket.Address,
		RelativeAmount:  bucket.RelativeAmount,
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
