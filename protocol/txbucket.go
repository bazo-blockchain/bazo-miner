package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"sort"
)

type TxBucket struct {
	Address         [64]byte
	RelativeBalance int64

	merkleRoot HashType // should not be accessed directly, instead, call CalculateMerkleRoot()

	Transactions []*FundsTx // won't be serialized
}

func NewTxBucket(address [64]byte) *TxBucket {
	return &TxBucket{
		Address:	address,
	}
}

func (bucket *TxBucket) AddFundsTx(tx *FundsTx) {
	if tx.From == bucket.Address {
		// Bucket owner is sender of the transaction, must pay the amount and the fee
		bucket.RelativeBalance -= int64(tx.Amount + tx.Fee)
	} else if tx.To == bucket.Address {
		// Bucket owner is receiver of the transaction, must not pay for the fee
		bucket.RelativeBalance += int64(tx.Amount)
	} else {
		return
	}

	bucket.Transactions = append(bucket.Transactions, tx)
}

func (bucket *TxBucket) CalculateMerkleRoot() HashType {
	emptyHash := HashType{}
	if bucket.merkleRoot == emptyHash {
		merkleTree := bucket.buildMerkleTree()
		bucket.merkleRoot = merkleTree.MerkleRoot()
	}
	return bucket.merkleRoot
}

func (bucket *TxBucket) buildMerkleTree() *MerkleTree {
	if bucket == nil {
		return nil
	}

	//Merkle root for no transactions is 0 hash
	if len(bucket.Transactions) == 0 {
		return nil
	}

	var txHashes HashArray
	for _, tx := range bucket.Transactions {
		txHashes = append(txHashes, tx.Hash())
	}

	m, _ := NewMerkleTree(txHashes)

	return m
}

func (bucket *TxBucket) Hash() HashType {
	if bucket == nil {
		return HashType{}
	}

	bucketHash := struct {
		address    [64]byte
		amount     int64
		merkleRoot HashType
	}{
		bucket.Address,
		bucket.RelativeBalance,
		bucket.CalculateMerkleRoot(),
	}

	return SerializeHashContent(bucketHash)
}

func (bucket *TxBucket) Encode() []byte {
	if bucket == nil {
		return nil
	}

	encoded := TxBucket{
		Address:         bucket.Address,
		RelativeBalance: bucket.RelativeBalance,
		merkleRoot:      bucket.CalculateMerkleRoot(),
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
	hash := bucket.Hash()
	merkleRoot := bucket.CalculateMerkleRoot()

	return fmt.Sprintf("\nHash: %x\n" +
		"Address: %x\n" +
		"Relative Balance: %v\n"+
		"Merkle Root: %x\n",
		hash[0:8],
		bucket.Address[0:8],
		bucket.RelativeBalance,
		merkleRoot[0:8],
	)
}

type txBucketMap map[AddressType]*TxBucket

func NewTxBucketMap() txBucketMap {
	return make(txBucketMap)
}

func (m txBucketMap) Sort() txBucketMap {
	var buckets []*TxBucket
	for _, bucket := range m {
		buckets = append(buckets, bucket)
	}

	sort.Slice(buckets, func(i, j int) bool {
		switch bytes.Compare(buckets[i].Address[:], buckets[j].Address[:]) {
		case -1:
			return true
		case 0, 1:
			return false
		default:
			log.Panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
			return false
		}
	})

	sortedMap := NewTxBucketMap()
	for _, bucket := range buckets {
		sortedMap[bucket.Address] = bucket
	}

	return sortedMap
}