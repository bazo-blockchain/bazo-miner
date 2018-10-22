package protocol

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"crypto/rsa"
)

const (
	STAKETX_SIZE = 394
)

//when we broadcast transactions we need a way to distinguish with a type

type StakeTx struct {
	Header     byte      // 1 Byte
	Fee        uint64    // 8 Byte
	IsStaking  bool      // 1 Byte
	HashedSeed [32]byte  // 32 Byte
	Account    [32]byte  // 32 Byte
	Sig        [64]byte  // 64 Byte
	CommKey    [256]byte // 256 Byte, RSA Public Key
}

func ConstrStakeTx(header byte, fee uint64, isStaking bool, hashedSeed, account [32]byte, signKey *ecdsa.PrivateKey, commKey *rsa.PublicKey) (tx *StakeTx, err error) {

	tx = new(StakeTx)

	tx.Header = header
	tx.Fee = fee
	tx.IsStaking = isStaking
	tx.HashedSeed = hashedSeed
	tx.Account = account

	txHash := tx.Hash()

	r, s, err := ecdsa.Sign(rand.Reader, signKey, txHash[:])
	if err != nil {
		return nil, err
	}

	copy(tx.Sig[32-len(r.Bytes()):32], r.Bytes())
	copy(tx.Sig[64-len(s.Bytes()):], s.Bytes())
	copy(tx.CommKey[256-len(commKey.N.Bytes()):], commKey.N.Bytes())

	return tx, nil
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
		CommKey    [256]byte
	}{
		tx.Header,
		tx.Fee,
		tx.IsStaking,
		tx.HashedSeed,
		tx.Account,
		tx.CommKey,
	}

	return SerializeHashContent(txHash)
}

//when we serialize the struct with binary.Write, unexported field get serialized as well, undesired
//behavior. Therefore, writing own encoder/decoder
func (tx *StakeTx) Encode() (encodedTx []byte) {
	if tx == nil {
		return nil
	}

	var fee [8]byte
	var isStaking byte

	binary.BigEndian.PutUint64(fee[:], tx.Fee)

	if tx.IsStaking == true {
		isStaking = 1
	} else {
		isStaking = 0
	}

	encodedTx = make([]byte, STAKETX_SIZE)

	encodedTx[0] = tx.Header
	copy(encodedTx[1:9], fee[:])
	encodedTx[9] = isStaking
	copy(encodedTx[10:42], tx.HashedSeed[:])
	copy(encodedTx[42:74], tx.Account[:])
	copy(encodedTx[74:138], tx.Sig[:])
	copy(encodedTx[138:394], tx.CommKey[:])

	return encodedTx
}

func (*StakeTx) Decode(encodedTx []byte) (tx *StakeTx) {
	tx = new(StakeTx)

	if len(encodedTx) != STAKETX_SIZE {
		return nil
	}

	var isStakingAsByte byte

	tx.Header = encodedTx[0]
	tx.Fee = binary.BigEndian.Uint64(encodedTx[1:9])
	isStakingAsByte = encodedTx[9]
	copy(tx.HashedSeed[:], encodedTx[10:42])
	copy(tx.Account[:], encodedTx[42:74])
	copy(tx.Sig[:], encodedTx[74:138])
	copy(tx.CommKey[:], encodedTx[138:394])

	if isStakingAsByte == 0 {
		tx.IsStaking = false
	} else {
		tx.IsStaking = true
	}

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
			"Sig: %x\n"+
			"Comm: %x\n",
		tx.Header,
		tx.Fee,
		tx.IsStaking,
		tx.HashedSeed[0:8],
		tx.Account[0:8],
		tx.Sig[0:8],
		tx.CommKey[0:8],
	)
}
