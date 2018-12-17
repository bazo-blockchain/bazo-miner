package protocol

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestBucketCreation(t *testing.T) {
	var address [64]byte
	rand.Read(address[:])
	bucket := NewTxBucket(address)

	if !reflect.DeepEqual(bucket.Address, address) {
		t.Errorf("BucketAddress does not match the given one: %x vs. %x", bucket.Address, address)
	}
}

func TestBucketHash(t *testing.T) {
	var address [64]byte
	rand.Read(address[:])
	bucket := NewTxBucket(address)

	hash1 := bucket.Hash()

	if !reflect.DeepEqual(hash1, bucket.Hash()) {
		t.Errorf("Bucket hashing failed!")
	}

	rand.Read(address[:])
	bucket.Address = address

	hash2 := bucket.Hash()

	if reflect.DeepEqual(hash1, hash2) {
		t.Errorf("Bucket hashing failed!")
	}

	if !reflect.DeepEqual(hash2, bucket.Hash()) {
		t.Errorf("Bucket hashing failed!")
	}
}

func TestBucketSerialization(t *testing.T) {
	var bucket TxBucket

	rand.Read(bucket.Address[:])
	bucket.RelativeBalance = rand.Int63()

	encodedBucket := bucket.Encode()

	var compareBucket TxBucket
	compareBucket = *compareBucket.Decode(encodedBucket)
	if !reflect.DeepEqual(bucket, compareBucket) {
		t.Error("Bucket encoding/decoding failed!")
	}
}

func TestBucketAddFundsTx(t *testing.T) {
	bucket := NewTxBucket(accA.Address)

	amount := uint64(100)
	fee := uint64(1)

	tx := NewSimpleFundsTx(amount, fee, uint32(0), accA.Address, accB.Address)
	bucket.AddFundsTx(tx)

	expectedAmount := int64(amount + fee) * -1
	if bucket.RelativeBalance != expectedAmount {
		t.Errorf("invalid bucket amount, is %v but should be %v", bucket.RelativeBalance, expectedAmount)
	}

	amount = uint64(300)

	tx = NewSimpleFundsTx(amount, fee, uint32(0), accB.Address, accA.Address)
	bucket.AddFundsTx(tx)

	expectedAmount = expectedAmount + int64(amount)
	if bucket.RelativeBalance != expectedAmount {
		t.Errorf("invalid bucket amount, is %v but should be %v", bucket.RelativeBalance, expectedAmount)
	}

	var randomAddressA, randomAddressB [64]byte
	rand.Read(randomAddressA[:])
	rand.Read(randomAddressB[:])

	tx = NewSimpleFundsTx(amount, fee, uint32(0), randomAddressA, randomAddressB)
	bucket.AddFundsTx(tx)

	if bucket.RelativeBalance != expectedAmount {
		t.Errorf("invalid bucket amount, is %v but should be %v", bucket.RelativeBalance, expectedAmount)
	}
}