package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	CONSOLIDATIONTX_SIZE = 213
)

//when we broadcast transactions we need a way to distinguish with a type

type ConsolidationTx struct {
	Header    byte
	Account   [32]byte
}

func ConstrConsolidationTx(header byte) (tx *ConsolidationTx, err error) {
	tx = new(ConsolidationTx)
	tx.Header = header

	return tx, nil
}

func (tx *ConsolidationTx) Hash() (hash [32]byte) {
	if tx == nil {
		//is returning nil better?
		return [32]byte{}
	}

	txHash := struct {
		Header byte

	}{
		tx.Header,
	}

	return SerializeHashContent(txHash)
}

//when we serialize the struct with binary.Write, unexported field get serialized as well, undesired
//behavior. Therefore, writing own encoder/decoder
func (tx *ConsolidationTx) Encode() (encodedTx []byte) {
	if tx == nil {
		return nil
	}

	var amount, fee [8]byte
	var txCnt [4]byte


	encodedTx = make([]byte, CONSOLIDATIONTX_SIZE)

	encodedTx[0] = tx.Header
	copy(encodedTx[1:9], amount[:])
	copy(encodedTx[9:17], fee[:])
	copy(encodedTx[17:21], txCnt[:])

	return encodedTx
}

func (*ConsolidationTx) Decode(encodedTx []byte) (tx *FundsTx) {
	tx = new(FundsTx)

	if len(encodedTx) != CONFIGTX_SIZE {
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

func (tx *ConsolidationTx) Size() uint64  { return CONSOLIDATIONTX_SIZE }

func (tx ConsolidationTx) String() string {
	return fmt.Sprintf(
		"\nHeader: %v\n",
		tx.Header,
	)
}
