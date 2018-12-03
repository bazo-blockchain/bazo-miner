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
	MHashes	[][33]byte

	// Proof properties
	// Must equal the hashed data of FundsTx.Hash()
	PHeader byte
	PAmount uint64
	PFee    uint64
	PTxCnt  uint32
	PFrom   [64]byte
	PTo     [64]byte
	PData   []byte
}


func NewMerkleProof(height uint32, mhashes [][33]byte, header byte, amount uint64, fee uint64, txcnt uint32, from [64]byte, to [64]byte, data []byte) (proof MerkleProof) {
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
		MHashes	[][33]byte
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
		"Proof Header: %v\n" +
		"Proof Amount: %v\n"+
		"Proof Fee: %v\n"+
		"Proof TxCnt: %v\n"+
		"Proof From: %x\n"+
		"Proof To: %x\n"+
		"Proof Data: %v\n",
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

func (proof *MerkleProof) CalculateMerkleRoot() (computedHash [32]byte, err error) {
	phash := proof.getProofHash()

	computedHash = phash
	for _, mhash := range proof.MHashes {
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