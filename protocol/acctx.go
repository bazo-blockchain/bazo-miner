package protocol

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

const (
	ACCTX_SIZE = 169
)

type AccTx struct {
	Header byte
	Issuer [32]byte
	Fee    uint64
	PubKey [64]byte
	Sig    [64]byte
}

func ConstrAccTx(header byte, fee uint64, rootPrivKey *ecdsa.PrivateKey) (tx *AccTx, privKey *ecdsa.PrivateKey, err error) {

	tx = new(AccTx)
	tx.Header = header
	tx.Fee = fee

	newAccAddress, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	newAccPub1, newAccPub2 := newAccAddress.PublicKey.X.Bytes(), newAccAddress.PublicKey.Y.Bytes()
	copy(tx.PubKey[32-len(newAccPub1):32], newAccPub1)
	copy(tx.PubKey[64-len(newAccPub2):], newAccPub2)

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
		Header byte
		Issuer [32]byte
		Fee    uint64
		PubKey [64]byte
	}{
		tx.Header,
		tx.Issuer,
		tx.Fee,
		tx.PubKey,
	}
	return SerializeHashContent(txHash)
}

func (tx *AccTx) Encode() (encodedTx []byte) {

	if tx == nil {
		return nil
	}

	var buf bytes.Buffer
	var feeBuf [8]byte

	binary.Write(&buf, binary.BigEndian, tx.Fee)
	copy(feeBuf[:], buf.Bytes())

	encodedTx = make([]byte, ACCTX_SIZE)
	encodedTx[0] = tx.Header
	copy(encodedTx[1:33], tx.Issuer[:])
	copy(encodedTx[33:41], feeBuf[:])
	copy(encodedTx[41:105], tx.PubKey[:])
	copy(encodedTx[105:169], tx.Sig[:])

	return encodedTx
}

func (*AccTx) Decode(encodedTx []byte) (tx *AccTx) {

	if len(encodedTx) != ACCTX_SIZE {
		return nil
	}

	tx = new(AccTx)
	tx.Header = encodedTx[0]
	copy(tx.Issuer[:], encodedTx[1:33])
	tx.Fee = binary.BigEndian.Uint64(encodedTx[33:41])
	copy(tx.PubKey[:], encodedTx[41:105])
	copy(tx.Sig[:], encodedTx[105:169])

	return tx
}

func (tx *AccTx) TxFee() uint64 { return tx.Fee }
func (tx *AccTx) Size() uint64  { return ACCTX_SIZE }

func (tx AccTx) String() string {
	return fmt.Sprintf(
		"\n"+
			"Issuer: %x\n"+
			"Fee: %v\n"+
			"PubKey: %x\n"+
			"Sig: %x\n",
		tx.Issuer[0:8],
		tx.Fee,
		tx.PubKey[0:8],
		tx.Sig[0:8],
	)
}
