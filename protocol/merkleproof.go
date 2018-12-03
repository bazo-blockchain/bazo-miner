package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"golang.org/x/crypto/sha3"
)

type MerkleProof struct {
	// Proof height
	Height 	uint32

	// Merkle hashes
	// Intermediate hashes required to create a Merkle proof
	MHashes	[][32]byte

	// Proof properties
	// Must equal the hashed data of FundsTx.Hash()
	PHeader  byte
	PAmount uint64
	PFee    uint64
	PTxCnt  uint32
	PFrom   [64]byte
	PTo     [64]byte
	PData   []byte
}


func NewMerkleProof(height uint32, mhashes [][32]byte, header byte, amount uint64, fee uint64, txcnt uint32, from [64]byte, to [64]byte, data []byte) (proof MerkleProof) {
	proof.Height = height
	proof.MHashes = mhashes
	proof.PHeader = header
	proof.PAmount = amount
	proof.PFee = fee
	proof.PTxCnt = txcnt
	proof.PFrom = from
	proof.PTo = to
	proof.PData = data

	return proof
}

func (proof *MerkleProof) Hash() (hash [32]byte) {
	if proof == nil {
		return [32]byte{}
	}

	input := struct {
		Height 	uint32
		MHashes	[][32]byte
		PHeader byte
		PAmount uint64
		PFee    uint64
		PTxCnt  uint32
		PFrom   [64]byte
		PTo     [64]byte
		PData   []byte
	}{
		proof.Height,
		proof.MHashes,
		proof.PHeader,
		proof.PAmount,
		proof.PFee,
		proof.PTxCnt,
		proof.PFrom,
		proof.PTo,
		proof.PData,
	}

	return SerializeHashContent(input)
}

func (proof *MerkleProof) Encode() (encodedTx []byte) {
	encodeData := MerkleProof{
		proof.Height,
		proof.MHashes,
		proof.PHeader,
		proof.PAmount,
		proof.PFee,
		proof.PTxCnt,
		proof.PFrom,
		proof.PTo,
		proof.PData,
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
	for _, mhash := range proof.MHashes {
		mhashesString += fmt.Sprintf("%x, ", mhash[0:8])
	}

	mhashesString = mhashesString[0:len(mhashesString) - 2]

	return fmt.Sprintf("Height: %v\n" +
		"MHashes: [%v]\n" +
		"PHash Header: %v\n" +
		"PHash Amount: %v\n"+
		"PHash Fee: %v\n"+
		"PHash TxCnt: %v\n"+
		"PHash From: %x\n"+
		"PHash To: %x\n"+
		"PHash Data: %v\n",
		proof.Height,
		mhashesString,
		proof.PHeader,
		proof.PAmount,
		proof.PFee,
		proof.PTxCnt,
		proof.PFrom[0:8],
		proof.PTo[0:8],
		proof.PData,
	)
}

func (proof *MerkleProof) CalculateMerkleRoot() (merkleRoot [32]byte, err error) {
	phash := proof.getProofHash()

	merkleRoot = phash
	for _, mhash := range proof.MHashes {
		concatHash := append(merkleRoot[:], mhash[:]...)
		merkleRoot = sha3.Sum256(concatHash)
	}

	return merkleRoot, nil
}

func (proof *MerkleProof) getProofHash() [32]byte {
	// Note that the hashed properties must equal to the hashed properties of FundsTx.Hash()
	hash := struct {
		Header byte
		Amount uint64
		Fee    uint64
		TxCnt  uint32
		From   [64]byte
		To     [64]byte
		Data   []byte
	}{
		proof.PHeader,
		proof.PAmount,
		proof.PFee,
		proof.PTxCnt,
		proof.PFrom,
		proof.PTo,
		proof.PData,
	}

	return SerializeHashContent(hash)
}