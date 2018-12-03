package protocol

import (
	"bytes"
	"encoding/gob"
)

type SCP struct {
	Proofs []*MerkleProof
}

func NewSCP() SCP {
	return SCP{}
}

func (scp *SCP) AddMerkleProof(proof *MerkleProof) (index int) {
	scp.Proofs = append(scp.Proofs, proof)
	return len(scp.Proofs) - 1
}

func (scp *SCP) CalculateMerkleRootFor(index int) (merkleRoot [32]byte, err error) {
	return scp.Proofs[index].CalculateMerkleRoot()

}

func (scp *SCP) ProofCount() int {
	return len(scp.Proofs)
}

func (scp *SCP) Hash() (hash [32]byte) {
	if scp == nil {
		return [32]byte{}
	}

	var proofHashes [][32]byte
	for _, proof := range scp.Proofs {
		proofHashes = append(proofHashes, proof.Hash())
	}

	return SerializeHashContent(proofHashes)
}

func (scp *SCP) Encode() (encodedTx []byte) {
	gob.Register(MerkleProof{})

	encodeData := SCP{
		scp.Proofs,
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
	return scp.Proofs[index].String()
}
