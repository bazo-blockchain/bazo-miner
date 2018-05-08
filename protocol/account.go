package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"
)

const (
	ACC_SIZE = 113
)

type Account struct {
	Address            [64]byte  // 64 Byte
	Balance            uint64    // 8 Byte
	TxCnt              uint32    // 4 Byte
	IsStaking          bool      // 1 Byte
	HashedSeed         [32]byte  // 32 Byte
	StakingBlockHeight uint32    // 4 Byte
	Contract           []byte    // Arbitrary length
	ContractVariables  []big.Int // Arbitrary length
}

func NewAccount(address [64]byte, balance uint64, isStaking bool, hashedSeed [32]byte) Account {
	newAcc := Account{
		address,
		balance,
		0,
		isStaking,
		hashedSeed,
		0,
		[]byte{},
		[]big.Int{},
	}

	return newAcc
}

func (acc *Account) Hash() (hash [32]byte) {
	if acc == nil {
		return [32]byte{}
	}

	return SerializeHashContent(acc.Address)
}

func (acc *Account) Encode() (encodedAcc []byte) {

	if acc == nil {
		return nil
	}

	encodedAcc = make([]byte, ACC_SIZE)

	var balanceBuf [8]byte
	var txCntBuf [4]byte
	var isStakingBuf [1]byte
	var stakingBlockHeightBuf [4]byte

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, acc.IsStaking)
	copy(isStakingBuf[:], buf.Bytes())

	binary.BigEndian.PutUint64(balanceBuf[:], acc.Balance)
	binary.BigEndian.PutUint32(txCntBuf[:], acc.TxCnt)
	binary.BigEndian.PutUint32(stakingBlockHeightBuf[:], acc.StakingBlockHeight)

	copy(encodedAcc[0:64], acc.Address[:])
	copy(encodedAcc[64:72], balanceBuf[:])
	copy(encodedAcc[72:76], txCntBuf[:])
	copy(encodedAcc[76:77], isStakingBuf[:])
	copy(encodedAcc[77:109], acc.HashedSeed[:])
	copy(encodedAcc[109:113], stakingBlockHeightBuf[:])

	return encodedAcc
}

func (*Account) Decode(encodedAcc []byte) (acc *Account) {

	if len(encodedAcc) != ACC_SIZE {
		return nil
	}

	acc = new(Account)
	copy(acc.Address[:], encodedAcc[0:64])
	acc.Balance = binary.BigEndian.Uint64(encodedAcc[64:72])
	acc.TxCnt = binary.BigEndian.Uint32(encodedAcc[72:76])
	acc.IsStaking, _ = strconv.ParseBool(string(encodedAcc[76:77]))
	copy(acc.HashedSeed[:], encodedAcc[77:109])
	acc.StakingBlockHeight = binary.BigEndian.Uint32(encodedAcc[109:113])

	return acc
}

func (acc Account) String() string {
	addressHash := SerializeHashContent(acc.Address)
	return fmt.Sprintf("Hash: %x, Address: %x, TxCnt: %v, Balance: %v, IsStaking: %v, HashedSeed: %x, StakingBlockHeight: %v", addressHash[0:8], acc.Address[0:8], acc.TxCnt, acc.Balance, acc.IsStaking, acc.HashedSeed[0:8], acc.StakingBlockHeight)
}
