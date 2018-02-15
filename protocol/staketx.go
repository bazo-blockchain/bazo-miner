package protocol

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

const (
	STAKETX_SIZE = 138
)

//when we broadcast transactions we need a way to distinguish with a type

type StakeTx struct {
	Header     byte     // 1 Byte
	Fee        uint64   // 8 Byte
	IsStaking  bool     // 1 Byte
	HashedSeed [32]byte // 32 Byte
	Account    [32]byte // 32 Byte
	Sig        [64]byte // 64 Byte
}

func ConstrStakeTx(header byte, fee uint64, isStaking bool, hashedSeed, account [32]byte, key *ecdsa.PrivateKey) (tx *StakeTx, err error) {

	tx = new(StakeTx)

	tx.Header = header
	tx.Fee = fee
	tx.IsStaking = isStaking
	tx.HashedSeed = hashedSeed
	tx.Account = account
	txHash := tx.Hash()

	r, s, err := ecdsa.Sign(rand.Reader, key, txHash[:])

	copy(tx.Sig[32-len(r.Bytes()):32], r.Bytes())
	copy(tx.Sig[64-len(s.Bytes()):], s.Bytes())
	return
}

func (tx *StakeTx) Hash() (hash [32]byte) {

	if tx == nil {
		//is returning nil better?
		return [32]byte{}
	}

	txHash := struct {
		Header     byte
		Fee        uint64
		IsStaking  bool
		hashedSeed [32]byte
		Account    [32]byte
	}{
		tx.Header,
		tx.Fee,
		tx.IsStaking,
		tx.HashedSeed,
		tx.Account,
	}
	return SerializeHashContent(txHash)
}

//when we serialize the struct with binary.Write, unexported field get serialized as well, undesired
//behavior. Therefore, writing own encoder/decoder
func (tx *StakeTx) Encode() (encodedTx []byte) {

	if tx == nil {
		return nil
	}

	var buf bytes.Buffer
	var feeBuf [8]byte
	var isStakingBuf [1]byte

	//transfer integer values to byte arrays
	binary.Write(&buf, binary.BigEndian, tx.Fee)
	copy(feeBuf[:], buf.Bytes())
	buf.Reset()
	binary.Write(&buf, binary.BigEndian, tx.IsStaking)
	copy(isStakingBuf[:], buf.Bytes())
	//fmt.Println("StakingBuf: ", isStakingBuf)
	buf.Reset()

	//fmt.Println("\n\nENCODING Hashed Secret: ", tx.HashedSeed)
	//fmt.Println("\n\nEnCODING Hashed PubKey: ", tx.Account)

	encodedTx = make([]byte, STAKETX_SIZE)
	encodedTx[0] = tx.Header
	copy(encodedTx[1:9], feeBuf[:])
	copy(encodedTx[9:10], isStakingBuf[:])
	copy(encodedTx[10:42], tx.HashedSeed[:])
	copy(encodedTx[42:74], tx.Account[:])
	copy(encodedTx[74:138], tx.Sig[:])
	//fmt.Println("encoded: ", encodedTx)

	return encodedTx
}

func (*StakeTx) Decode(encodedTx []byte) (tx *StakeTx) {

	if len(encodedTx) != STAKETX_SIZE {
		return nil
	}

	tx = new(StakeTx)
	tx.Header = encodedTx[0]

	tx.Fee = binary.BigEndian.Uint64(encodedTx[1:9])
	tx.IsStaking = encodedTx[9:10][0] != 0
	copy(tx.HashedSeed[:], encodedTx[10:42])
	copy(tx.Account[:], encodedTx[42:74])
	copy(tx.Sig[:], encodedTx[74:138])

	//fmt.Println("\n\nDECODING: ", encodedTx)
	return tx
}

func (tx *StakeTx) TxFee() uint64 { return tx.Fee }
func (tx *StakeTx) Size() uint64  { return STAKETX_SIZE }

func (tx StakeTx) String() string {
	return fmt.Sprintf(
		"\nHeader: %x\n"+
			"Fee: %v\n"+
			"IsStaking: %v\n"+
			"hashedSeed: %x\n"+
			"Account: %x\n"+
			"Sig: %x\n",
		tx.Header,
		tx.Fee,
		tx.IsStaking,
		tx.HashedSeed[0:8],
		tx.Account[0:8],
		tx.Sig[0:8],
	)
}
