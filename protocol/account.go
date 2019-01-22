package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
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

func (acc *Account) Hash() [32]byte {
	if acc == nil {
		return [32]byte{}
	}

	return SerializeHashContent(acc.Address)
}

func (acc *Account) Encode() []byte {
	if acc == nil {
		return nil
	}

	encoded := Account{
		Address:            acc.Address,
		Balance:            acc.Balance,
		TxCnt:              acc.TxCnt,
		IsStaking:          acc.IsStaking,
		HashedSeed:         acc.HashedSeed,
		StakingBlockHeight: acc.StakingBlockHeight,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (*Account) Decode(encoded []byte) (acc *Account) {
	var decoded Account
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (acc Account) String() string {
	addressHash := acc.Hash()
	return fmt.Sprintf("Hash: %x, Address: %x, TxCnt: %v, Balance: %v, IsStaking: %v, HashedSeed: %x, StakingBlockHeight: %v", addressHash[0:8], acc.Address[0:8], acc.TxCnt, acc.Balance, acc.IsStaking, acc.HashedSeed[0:8], acc.StakingBlockHeight)
}
