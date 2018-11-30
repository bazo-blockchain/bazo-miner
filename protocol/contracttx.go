package protocol

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"fmt"
)

const (
	CONTRACTTX_SIZE = 201
)

type ContractTx struct {
	Header            byte
	Issuer            [64]byte
	Fee               uint64
	PubKey            [64]byte
	Sig               [64]byte
	Contract          []byte
	ContractVariables []ByteArray
}

func ConstrContractTx(header byte, fee uint64, issuerSigKey *ecdsa.PrivateKey, contract []byte, contractVariables []ByteArray) (tx *ContractTx, newContractKey *ecdsa.PrivateKey, err error) {
	tx = new(ContractTx)
	tx.Header = header
	tx.Fee = fee
	tx.Contract = contract
	tx.ContractVariables = contractVariables

	newContractKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	newAccPub1, newAccPub2 := newContractKey.PublicKey.X.Bytes(), newContractKey.PublicKey.Y.Bytes()
	copy(tx.PubKey[32-len(newAccPub1):32], newAccPub1)
	copy(tx.PubKey[64-len(newAccPub2):], newAccPub2)

	var issuerPublicKey [64]byte
	issuerPubKey1, issuerPubKey2 := issuerSigKey.PublicKey.X.Bytes(), issuerSigKey.PublicKey.Y.Bytes()
	copy(issuerPublicKey[32-len(issuerPubKey1):32], issuerPubKey1)
	copy(issuerPublicKey[64-len(issuerPubKey2):], issuerPubKey2)
	copy(tx.Issuer[:], issuerPublicKey[:])

	txHash := tx.Hash()

	r, s, err := ecdsa.Sign(rand.Reader, issuerSigKey, txHash[:])
	if err != nil {
		return nil, nil, err
	}

	copy(tx.Sig[32-len(r.Bytes()):32], r.Bytes())
	copy(tx.Sig[64-len(s.Bytes()):], s.Bytes())

	return tx, newContractKey, nil
}

func (tx *ContractTx) Hash() [32]byte {
	if tx == nil {
		return [32]byte{}
	}

	txHash := struct {
		Header            byte
		Issuer            [64]byte
		Fee               uint64
		PubKey            [64]byte
		Contract          []byte
		ContractVariables []ByteArray
	}{
		tx.Header,
		tx.Issuer,
		tx.Fee,
		tx.PubKey,
		tx.Contract,
		tx.ContractVariables,
	}

	return SerializeHashContent(txHash)
}

func (tx *ContractTx) Encode() []byte {
	if tx == nil {
		return nil
	}

	encoded := ContractTx{
		Header: tx.Header,
		Issuer: tx.Issuer,
		Fee:    tx.Fee,
		PubKey: tx.PubKey,
		Sig:    tx.Sig,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (*ContractTx) Decode(encoded []byte) (tx *ContractTx) {
	var decoded ContractTx
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (tx *ContractTx) TxFee() uint64 { return tx.Fee }

func (tx *ContractTx) Size() uint64 { return CONTRACTTX_SIZE }

func (tx ContractTx) String() string {
	return fmt.Sprintf(
		"\n"+
			"Header: %x\n"+
			"Issuer: %x\n"+
			"Fee: %v\n"+
			"PubKey: %x\n"+
			"Sig: %x\n"+
			"Contract: %v\n"+
			"ContractVariables: %v\n",
		tx.Header,
		tx.Issuer[0:8],
		tx.Fee,
		tx.PubKey[0:8],
		tx.Sig[0:8],
		tx.Contract[:],
		tx.ContractVariables[:],
	)
}
