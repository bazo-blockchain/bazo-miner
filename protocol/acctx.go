package protocol

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"math/big"
)

const (
	ACCTX_SIZE = 169
)

type AccTx struct {
	Header            byte
	Issuer            [32]byte
	Fee               uint64
	PubKey            [64]byte
	Sig               [64]byte
	Contract          []byte
	ContractVariables []big.Int
}

func ConstrAccTx(header byte, fee uint64, address [64]byte, rootPrivKey *ecdsa.PrivateKey, contract []byte, contractVariables []big.Int) (tx *AccTx, newAccAddress *ecdsa.PrivateKey, err error) {
	tx = new(AccTx)
	tx.Header = header
	tx.Fee = fee
	tx.Contract = contract
	tx.ContractVariables = contractVariables

	var newAccAddressString string

	if address != [64]byte{} {
		copy(tx.PubKey[:], address[:])
	} else {
		//Check if string representation of account address is 128 long. Else there will be problems when doing REST calls.
		for len(newAccAddressString) != 128 {
			newAccAddress, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			newAccPub1, newAccPub2 := newAccAddress.PublicKey.X.Bytes(), newAccAddress.PublicKey.Y.Bytes()
			copy(tx.PubKey[32-len(newAccPub1):32], newAccPub1)
			copy(tx.PubKey[64-len(newAccPub2):], newAccPub2)

			newAccAddressString = newAccAddress.X.Text(16) + newAccAddress.Y.Text(16)
		}
	}

	var rootPublicKey [64]byte
	rootPubKey1, rootPubKey2 := rootPrivKey.PublicKey.X.Bytes(), rootPrivKey.PublicKey.Y.Bytes()
	copy(rootPublicKey[32-len(rootPubKey1):32], rootPubKey1)
	copy(rootPublicKey[64-len(rootPubKey2):], rootPubKey2)

	issuer := SerializeHashContent(rootPublicKey)
	copy(tx.Issuer[:], issuer[:])

	txHash := tx.Hash()

	r, s, err := ecdsa.Sign(rand.Reader, rootPrivKey, txHash[:])
	if err != nil {
		return nil, nil, err
	}

	copy(tx.Sig[32-len(r.Bytes()):32], r.Bytes())
	copy(tx.Sig[64-len(s.Bytes()):], s.Bytes())

	return tx, newAccAddress, nil
}

func (tx *AccTx) Hash() (hash [32]byte) {

	if tx == nil {
		return [32]byte{}
	}

	txHash := struct {
		Header            byte
		Issuer            [32]byte
		Fee               uint64
		PubKey            [64]byte
		Contract          []byte
		ContractVariables []big.Int
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

func (tx *AccTx) Encode() (encodedTx []byte) {
	// Encode
	encodeData := AccTx{
		tx.Header,
		tx.Issuer,
		tx.Fee,
		tx.PubKey,
		tx.Sig,
		tx.Contract,
		tx.ContractVariables,
	}
	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encodeData)
	return buffer.Bytes()
}

func (*AccTx) Decode(encodedTx []byte) *AccTx {
	var decoded AccTx
	buffer := bytes.NewBuffer(encodedTx)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (tx *AccTx) TxFee() uint64 { return tx.Fee }

func (tx *AccTx) Size() uint64 { return ACCTX_SIZE }

func (tx AccTx) String() string {
	return fmt.Sprintf(
		"\n"+
			"Issuer: %x\n"+
			"Fee: %v\n"+
			"PubKey: %x\n"+
			"Sig: %x\n"+
			"Contract: %v\n"+
			"ContractVariables: %v\n",
		tx.Issuer[0:8],
		tx.Fee,
		tx.PubKey[0:8],
		tx.Sig[0:8],
		tx.Contract[:],
		tx.ContractVariables[:],
	)
}
