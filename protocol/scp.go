package protocol

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
)

type SCP struct {
	// Proof height
	// Block heights of the proofs
	Height 	[]uint32

	// Merkle hashes
	// Intermediate hashes required to create a Merkle proof
	MHashes	[][][32]byte

	// Proof properties
	// Must equal the hashed data of FundsTx.Hash()
	PHeader []byte
	PAmount []uint64
	PFee    []uint64
	PTxCnt  []uint32
	PFrom   [][64]byte
	PTo     [][64]byte
	PData   [][]byte
}

func (scp *SCP) CalculateMerkleRoot(index int) (merkleRoot [32]byte, err error) {
	if err = scp.verifyProofPrerequisites(); err != nil {
		return [32]byte{}, err
	}

	phash := scp.getProofHash(index)
	mhashes := scp.MHashes[index]

	merkleRoot = phash
	for _, mhash := range mhashes {
		concatHash := append(merkleRoot[:], mhash[:]...)
		merkleRoot = sha3.Sum256(concatHash)
	}

	return merkleRoot, nil
}

func (scp *SCP) ProofCount() int {
	return len(scp.Height)
}

func (scp *SCP) Hash() (hash [32]byte) {
	if scp == nil {
		return [32]byte{}
	}

	scpHash := struct {
		Height 		[]uint32
		MHashes		[][][32]byte
		PHashHeader []byte
		PHashAmount []uint64
		PHashFee    []uint64
		PHashTxCnt  []uint32
		PHashFrom   [][64]byte
		PHashTo     [][64]byte
		PHashData   [][]byte
	}{
		scp.Height,
		scp.MHashes,
		scp.PHeader,
		scp.PAmount,
		scp.PFee,
		scp.PTxCnt,
		scp.PFrom,
		scp.PTo,
		scp.PData,
	}

	return SerializeHashContent(scpHash)
}

func (scp *SCP) Encode() (encodedTx []byte) {
	encodeData := SCP{
		scp.Height,
		scp.MHashes,
		scp.PHeader,
		scp.PAmount,
		scp.PFee,
		scp.PTxCnt,
		scp.PFrom,
		scp.PTo,
		scp.PData,
	}
	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encodeData)
	return buffer.Bytes()
}

func (scp *SCP) Decode(encoded []byte) *SCP {
	var decoded SCP
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (scp *SCP) StringFor(index int) string {
	var mhashesString string
	for _, mhash := range scp.MHashes[index] {
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
		scp.Height[index],
		mhashesString,
		scp.PHeader[index],
		scp.PAmount[index],
		scp.PFee[index],
		scp.PTxCnt[index],
		scp.PFrom[index][0:8],
		scp.PTo[index][0:8],
		scp.PData[index],
	)
}

func (scp *SCP) getProofHash(index int) [32]byte {
	// Note that the hashed properties must equal to the hashed properties of FundsTx.Hash()
	phash := struct {
		Header byte
		Amount uint64
		Fee    uint64
		TxCnt  uint32
		From   [64]byte
		To     [64]byte
		Data   []byte
	}{
		scp.PHeader[index],
		scp.PAmount[index],
		scp.PFee[index],
		scp.PTxCnt[index],
		scp.PFrom[index],
		scp.PTo[index],
		scp.PData[index],
	}

	return SerializeHashContent(phash)
}

func (scp *SCP) verifyProofPrerequisites() error {
	nofProofs := scp.ProofCount()

	if nofProofs != len(scp.MHashes) {
		// For every proof i, the number of Merkle hashes at i must be specified.
		return errors.New("missing Merkle hashes")
	}

	if nofProofs != len(scp.PHeader) ||
		nofProofs != len(scp.PAmount) ||
		nofProofs != len(scp.PFee) ||
		nofProofs != len(scp.PTxCnt) ||
		nofProofs != len(scp.PFrom) ||
		nofProofs != len(scp.PTo) ||
		nofProofs != len(scp.PData) {
		// For every proof i, the properties for proof i must be specified.
		return errors.New("missing proof hash values")
	}

	return nil
}