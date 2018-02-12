package protocol

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

const (
	FUNDSTX_SIZE = 149
)

//when we broadcast transactions we need a way to distinguish with a type

type FundsTx struct {
	Header byte
	Amount uint64
	Fee    uint64
	TxCnt  uint32
	From   [32]byte
	To     [32]byte
	Sig    [64]byte
}

func ConstrFundsTx(header byte, amount uint64, fee uint64, txCnt uint32, from, to [32]byte, key *ecdsa.PrivateKey) (tx *FundsTx, err error) {
	tx = new(FundsTx)

	tx.From = from
	tx.To = to
	tx.Header = header
	tx.Amount = amount
	tx.Fee = fee
	tx.TxCnt = txCnt

	txHash := tx.Hash()

	r, s, err := ecdsa.Sign(rand.Reader, key, txHash[:])

	copy(tx.Sig[32-len(r.Bytes()):32], r.Bytes())
	copy(tx.Sig[64-len(s.Bytes()):], s.Bytes())

	return
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

	var buf bytes.Buffer
	var amountBuf [8]byte
	var feeBuf [8]byte
	var txCntBuf [4]byte

	//transfer integer values to byte arrays
	binary.Write(&buf, binary.BigEndian, tx.Amount)
	copy(amountBuf[:], buf.Bytes())
	buf.Reset()
	binary.Write(&buf, binary.BigEndian, tx.Fee)
	copy(feeBuf[:], buf.Bytes())
	buf.Reset()
	binary.Write(&buf, binary.BigEndian, tx.TxCnt)
	copy(txCntBuf[:], buf.Bytes())
	buf.Reset()

	encodedTx = make([]byte, FUNDSTX_SIZE)
	encodedTx[0] = tx.Header
	copy(encodedTx[1:9], amountBuf[:])
	copy(encodedTx[9:17], feeBuf[:])
	copy(encodedTx[17:21], txCntBuf[:])
	copy(encodedTx[21:53], tx.From[:])
	copy(encodedTx[53:85], tx.To[:])
	copy(encodedTx[85:149], tx.Sig[:])

	return encodedTx
}

func (*FundsTx) Decode(encodedTx []byte) (tx *FundsTx) {
	if len(encodedTx) != FUNDSTX_SIZE {
		return nil
	}

	tx = new(FundsTx)
	tx.Header = encodedTx[0]
	tx.Amount = binary.BigEndian.Uint64(encodedTx[1:9])
	tx.Fee = binary.BigEndian.Uint64(encodedTx[9:17])
	tx.TxCnt = binary.BigEndian.Uint32(encodedTx[17:21])
	copy(tx.From[:], encodedTx[21:53])
	copy(tx.To[:], encodedTx[53:85])
	copy(tx.Sig[:], encodedTx[85:149])

	return tx
}

func (tx *FundsTx) TxFee() uint64 { return tx.Fee }
func (tx *FundsTx) Size() uint64  { return FUNDSTX_SIZE }

func (tx FundsTx) String() string {
	return fmt.Sprintf(
		"\nHeader: %x\n"+
			"Amount: %v\n"+
			"Fee: %v\n"+
			"TxCnt: %v\n"+
			"From: %x\n"+
			"To: %x\n"+
			"Sig: %x\n",
		tx.Header,
		tx.Amount,
		tx.Fee,
		tx.TxCnt,
		tx.From[0:10],
		tx.To[0:10],
		tx.Sig[0:10],
	)
}
