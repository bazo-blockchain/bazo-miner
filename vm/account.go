package vm

import "math/big"

type ContractAccount struct {
	Address            [32]byte
	Balance            uint64
	TxCnt              uint64
	IsStaking          bool
	HashedSeed         [32]byte
	StakingBlockHeight uint64
	Contract           []byte    // Additional to standard account
	ContractVariables  []big.Int // Additional to standard account
}

func NewContractAccount(address [32]byte, balance uint64, isStaking bool, hashedSeed [32]byte, code []byte) ContractAccount {
	newSC := ContractAccount{
		address,
		balance,
		0,
		isStaking,
		hashedSeed,
		0,
		code,
		[]big.Int{},
	}
	return newSC
}

type ContractTx struct {
	Header          byte
	Amount          uint64
	Fee             int
	TxCnt           uint32
	From            [32]byte
	To              [32]byte
	TransactionData []byte
	Sig1            [64]byte
	Sig2            [64]byte
}
