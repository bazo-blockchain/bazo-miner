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
		t.Errorf("Address does not match the given one: %x vs. %x", bucket.Address, address)
	}
}

func TestBucketHash(t *testing.T) {
	var address [64]byte
	rand.Read(address[:])
	bucket := NewTxBucket(address)

	hash1 := bucket.CreateHash()

	if !reflect.DeepEqual(hash1, bucket.CreateHash()) {
		t.Errorf("Bucket hashing failed!")
	}

	rand.Read(address[:])
	bucket.Address = address

	hash2 := bucket.CreateHash()

	if reflect.DeepEqual(hash1, hash2) {
		t.Errorf("Bucket hashing failed!")
	}

	if !reflect.DeepEqual(hash2, bucket.CreateHash()) {
		t.Errorf("Bucket hashing failed!")
	}
}

func TestBucketSerialization(t *testing.T) {
	var bucket TxBucket

	rand.Read(bucket.Hash[:])
	rand.Read(bucket.Address[:])
	bucket.Amount = rand.Uint64()
	rand.Read(bucket.MerkleRoot[:])

	encodedBucket := bucket.Encode()

	var compareBucket TxBucket
	compareBucket = *compareBucket.Decode(encodedBucket)
	if !reflect.DeepEqual(bucket, compareBucket) {
		t.Error("Bucket encoding/decoding failed!")
	}
}