package protocol

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/gob"
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
	From   [64]byte
	To     [64]byte
	Sig    [64]byte
	Data   []byte
	//MPT_Proof *ethdb.MemDatabase
	MPT_Proof
}

func ConstrFundsTx(header byte, amount uint64, fee uint64, txCnt uint32, from, to [64]byte, sigKey *ecdsa.PrivateKey, data []byte) (tx *FundsTx, err error) {
	tx = new(FundsTx)
	tx.Header = header
	tx.From = from
	tx.To = to
	tx.Amount = amount
	tx.Fee = fee
	tx.TxCnt = txCnt
	tx.Data = data

	txHash := tx.Hash()

	r, s, err := ecdsa.Sign(rand.Reader, sigKey, txHash[:])
	if err != nil {
		return nil, err
	}

	copy(tx.Sig[32-len(r.Bytes()):32], r.Bytes())
	copy(tx.Sig[64-len(s.Bytes()):], s.Bytes())

	tx.MPT_Proof = nil

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
		From   [64]byte
		To     [64]byte
		Data   []byte
	}{
		tx.Header,
		tx.Amount,
		tx.Fee,
		tx.TxCnt,
		tx.From,
		tx.To,
		tx.Data,
	}

	return SerializeHashContent(txHash)
}

//when we serialize the struct with binary.Write, unexported field get serialized as well, undesired
//behavior. Therefore, writing own encoder/decoder
func (tx *FundsTx) Encode() (encodedTx []byte) {

	//gob.Register(&ethdb.MemDatabase{})

	// Encode
	encodeData := FundsTx{
		tx.Header,
		tx.Amount,
		tx.Fee,
		tx.TxCnt,
		tx.From,
		tx.To,
		tx.Sig,
		tx.Data,
		tx.MPT_Proof,
	}
	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encodeData)

	return buffer.Bytes()
}

func (*FundsTx) Decode(encodedTx []byte) *FundsTx {
	var decoded FundsTx
	buffer := bytes.NewBuffer(encodedTx)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
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
			"Sig: %x\n"+
			"Data: %v\n",
		tx.Header,
		tx.Amount,
		tx.Fee,
		tx.TxCnt,
		tx.From[0:8],
		tx.To[0:8],
		tx.Sig[0:8],
		tx.Data,
	)
}
