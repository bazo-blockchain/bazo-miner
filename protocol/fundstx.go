package protocol

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

const (
	FUNDSTX_SIZE = 213
)

//when we broadcast transactions we need a way to distinguish with a type

type FundsTx struct {
	Header byte
	Amount uint64
	Fee    uint64
	TxCnt  uint32
	From   [32]byte
	To     [32]byte
	Sig1   [64]byte
	Sig2   [64]byte
}

func ConstrFundsTx(header byte, amount uint64, fee uint64, txCnt uint32, from, to [32]byte, sig1Key *ecdsa.PrivateKey, sig2Key *ecdsa.PrivateKey) (tx *FundsTx, err error) {
	tx = new(FundsTx)

	tx.Header = header
	tx.From = from
	tx.To = to
	tx.Amount = amount
	tx.Fee = fee
	tx.TxCnt = txCnt

	txHash := tx.Hash()

	r, s, err := ecdsa.Sign(rand.Reader, sig1Key, txHash[:])
	if err != nil {
		return nil, err
	}

	copy(tx.Sig1[32-len(r.Bytes()):32], r.Bytes())
	copy(tx.Sig1[64-len(s.Bytes()):], s.Bytes())

	if sig2Key != nil {
		r, s, err := ecdsa.Sign(rand.Reader, sig2Key, txHash[:])
		if err != nil {
			return nil, err
		}

		copy(tx.Sig2[32-len(r.Bytes()):32], r.Bytes())
		copy(tx.Sig2[64-len(s.Bytes()):], s.Bytes())
	}

	return tx, nil
}

func (tx *FundsTx) Hash() (hash [32]byte) {
	if tx == nil {
		//is returning nil better?
		return [32]byte{}
	}

	txHash := struct {
		Header byte
		Amount uint64
		Fee    uint64
		TxCnt  uint32
		From   [32]byte
		To     [32]byte
	}{
		tx.Header,
		tx.Amount,
		tx.Fee,
		tx.TxCnt,
		tx.From,
		tx.To,
	}

	return SerializeHashContent(txHash)
}

//when we serialize the struct with binary.Write, unexported field get serialized as well, undesired
//behavior. Therefore, writing own encoder/decoder
func (tx *FundsTx) Encode() (encodedTx []byte) {
	if tx == nil {
		return nil
	}

	var amount, fee [8]byte
	var txCnt [4]byte

	binary.BigEndian.PutUint64(amount[:], tx.Amount)
	binary.BigEndian.PutUint64(fee[:], tx.Fee)
	binary.BigEndian.PutUint32(txCnt[:], tx.TxCnt)

	encodedTx = make([]byte, FUNDSTX_SIZE)

	encodedTx[0] = tx.Header
	copy(encodedTx[1:9], amount[:])
	copy(encodedTx[9:17], fee[:])
	copy(encodedTx[17:21], txCnt[:])
	copy(encodedTx[21:53], tx.From[:])
	copy(encodedTx[53:85], tx.To[:])
	copy(encodedTx[85:149], tx.Sig1[:])
	copy(encodedTx[149:213], tx.Sig2[:])

	return encodedTx
}

func (*FundsTx) Decode(encodedTx []byte) (tx *FundsTx) {
	tx = new(FundsTx)

	if len(encodedTx) != FUNDSTX_SIZE {
		return nil
	}

	tx.Header = encodedTx[0]
	tx.Amount = binary.BigEndian.Uint64(encodedTx[1:9])
	tx.Fee = binary.BigEndian.Uint64(encodedTx[9:17])
	tx.TxCnt = binary.BigEndian.Uint32(encodedTx[17:21])
	copy(tx.From[:], encodedTx[21:53])
	copy(tx.To[:], encodedTx[53:85])
	copy(tx.Sig1[:], encodedTx[85:149])
	copy(tx.Sig2[:], encodedTx[149:213])

	return tx
}

func (tx *FundsTx) TxFee() uint64 { return tx.Fee }
func (tx *FundsTx) Size() uint64  { return FUNDSTX_SIZE }

func (tx FundsTx) String() string {
	return fmt.Sprintf(
		"\nHeader: %v\n"+
			"Amount: %v\n"+
			"Fee: %v\n"+
			"TxCnt: %v\n"+
			"From: %x\n"+
			"To: %x\n"+
			"Sig1: %x\n"+
			"Sig2: %x\n",
		tx.Header,
		tx.Amount,
		tx.Fee,
		tx.TxCnt,
		tx.From[0:8],
		tx.To[0:8],
		tx.Sig1[0:8],
		tx.Sig2[0:8],
	)
}
