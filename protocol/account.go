package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	ACC_SIZE = 113
)

type Account struct {
	Address            [64]byte // 64 Byte
	Balance            uint64   // 8 Byte
	TxCnt              uint32   // 4 Byte
	IsStaking          bool     // 1 Byte
	HashedSeed         [32]byte // 32 Byte
	StakingBlockHeight uint32   // 4 Byte
}

func NewAccount(address [64]byte, balance uint64, isStaking bool, hashedSeed [32]byte) Account {
	newAcc := Account{
		address,
		balance,
		0,
		isStaking,
		hashedSeed,
		0,
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

	var balance [8]byte
	var txCnt [4]byte
	var isStaking byte
	var stakingBlockHeight [4]byte

	binary.BigEndian.PutUint64(balance[:], acc.Balance)
	binary.BigEndian.PutUint32(txCnt[:], acc.TxCnt)
	if acc.IsStaking {
		isStaking = 1
	}
	binary.BigEndian.PutUint32(stakingBlockHeight[:], acc.StakingBlockHeight)

	copy(encodedAcc[0:64], acc.Address[:])
	copy(encodedAcc[64:72], balance[:])
	copy(encodedAcc[72:76], txCnt[:])
	encodedAcc[76] = isStaking
	copy(encodedAcc[77:109], acc.HashedSeed[:])
	copy(encodedAcc[109:113], stakingBlockHeight[:])

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
	if encodedAcc[76] == 1 {
		acc.IsStaking = true
	} else {
		acc.IsStaking = false
	}
	copy(acc.HashedSeed[:], encodedAcc[77:109])
	acc.StakingBlockHeight = binary.BigEndian.Uint32(encodedAcc[109:113])

	return acc
}

func (acc Account) String() string {
	addressHash := SerializeHashContent(acc.Address)
	return fmt.Sprintf("Hash: %x, Address: %x, TxCnt: %v, Balance: %v, IsStaking: %v, HashedSeed: %x, StakingBlockHeight: %v", addressHash[0:8], acc.Address[0:8], acc.TxCnt, acc.Balance, acc.IsStaking, acc.HashedSeed[0:8], acc.StakingBlockHeight)
}
