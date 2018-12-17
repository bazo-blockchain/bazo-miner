package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type MerkleProof struct {
	// Proof height
	Height 	uint32

	// Merkle hashes
	// Intermediate hashes required to create a Merkle proof
	// Note that the first byte specifies the left or right node of the Merkle tree while the rest is the actual hash
	MerkleHashes [][33]byte

	// Bucket properties
	// Must equal the hashed data of TxBucket.Hash()
	BucketAddress         AddressType
	BucketRelativeBalance int64
	BucketMerkleRoot      HashType
}


func NewMerkleProof(height uint32, mhashes [][33]byte, address AddressType, amount int64, merkleRoot HashType) (proof MerkleProof) {
	proof.Height = height
	proof.MerkleHashes = mhashes
	proof.BucketAddress = address
	proof.BucketRelativeBalance = amount
	proof.BucketMerkleRoot = merkleRoot

	return proof
}

func (proof *MerkleProof) Hash() (hash [32]byte) {
	if proof == nil {
		return [32]byte{}
	}

	input := struct {
		Height     uint32
		MHashes    [][33]byte
		Address    AddressType
		Amount     int64
		MerkleRoot HashType
	}{
		proof.Height,
		proof.MerkleHashes,
		proof.BucketAddress,
		proof.BucketRelativeBalance,
		proof.BucketMerkleRoot,
	}

	return SerializeHashContent(input)
}

func (proof *MerkleProof) Encode() (encodedTx []byte) {
	encodeData := MerkleProof{
		proof.Height,
		proof.MerkleHashes,
		proof.BucketAddress,
		proof.BucketRelativeBalance,
		proof.BucketMerkleRoot,
	}
	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encodeData)
	return buffer.Bytes()
}

func (proof *MerkleProof) Decode(encoded []byte) *MerkleProof {
	var decoded MerkleProof
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (proof *MerkleProof) String() string {
	var mhashesString string
	for _, mhash := range proof.MerkleHashes {
		mhashesString += fmt.Sprintf("%x, ", mhash[0:8])
	}

	mhashesString = mhashesString[0:len(mhashesString) - 2]

	return fmt.Sprintf("Height: %v\n" +
		"MerkleHashes: [%v]\n" +
		"Bucket BucketAddress: %v\n" +
		"Bucket BucketRelativeBalance: %v\n"+
		"Bucket BucketMerkleRoot: %v\n",
		proof.Height,
		mhashesString,
		proof.BucketAddress[0:8],
		proof.BucketRelativeBalance,
		proof.BucketMerkleRoot[0:8],
	)
}

func (proof *MerkleProof) CalculateMerkleRoot() (computedHash [32]byte, err error) {
	bucketHash := proof.getBucketHash()

	computedHash = bucketHash
	for _, mhash := range proof.MerkleHashes {
		var hash [32]byte
		var leftOrRight [1]byte
		copy(leftOrRight[:], mhash[0:1])
		copy(hash[:], mhash[1:33])

		if leftOrRight == [1]byte{'l'} {
			computedHash = MTHash(append(hash[:], computedHash[:]...))
		} else {
			computedHash = MTHash(append(computedHash[:], hash[:]...))
		}
	}

	return computedHash, nil
}

func (proof *MerkleProof) getBucketHash() [32]byte {
	// Note that the hashed properties must equal to the hashed properties of TxBucket.Hash()
	input := struct {
		Address    AddressType
		Amount     int64
		MerkleRoot HashType
	}{
		proof.BucketAddress,
		proof.BucketRelativeBalance,
		proof.BucketMerkleRoot,
	}

	return SerializeHashContent(input)
}