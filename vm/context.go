package vm

import (
	"errors"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"math/big"
)

type StateData struct {
	data []byte
}

type Context struct {
	TransactionSender [32]byte
	TransactionData   []byte
	MaxGasAmount      uint64
	ContractAccount   ContractAccount
	ContractTx        ContractTx
	/*
		stateData StateData


		blockHeader []byte*/
}

func NewContext() *Context {
	//data := map[int][]byte{}

	return &Context{
		TransactionSender: [32]byte{},
		TransactionData:   []byte{},
		MaxGasAmount:      100000,
		ContractAccount:   ContractAccount{},
		ContractTx:        ContractTx{},
	}
}

type MockContext struct {
	protocol.Account
	changes []protocol.Change
	protocol.FundsTx
	transactionData []byte
}

func NewMockContext(byteCode []byte) *MockContext {
	mc := MockContext{}
	mc.SetContract(byteCode)
	mc.Fee = 50
	return &mc
}

func (mc *MockContext) SetContract(contract []byte) {
	mc.Contract = contract
}

func (mc *MockContext) GetContract() []byte {
	return mc.Contract
}

func (mc *MockContext) GetContractVariable(index int) (big.Int, error) {
	if index >= len(mc.ContractVariables) {
		return big.Int{}, errors.New("Index out of bounds")
	}
	return mc.ContractVariables[index], nil
}

func (mc *MockContext) SetContractVariable(index int, value big.Int) error {
	if len(mc.ContractVariables) <= index {
		return errors.New("Index out of bounds")
	}
	change := protocol.NewChange(index, value)
	mc.changes = append(mc.changes, change)
	return nil
}

func (mc *MockContext) PersistChanges() {
	for _, c := range mc.changes {
		i, v := c.GetChange()
		mc.ContractVariables[i] = v
	}
}

func (mc *MockContext) GetAddress() [64]byte {
	return mc.Address
}

func (mc *MockContext) GetBalance() uint64 {
	return mc.Balance
}

func (mc *MockContext) GetSender() [32]byte {
	return mc.From
}

func (mc *MockContext) GetAmount() uint64 {
	return mc.Amount
}

func (mc *MockContext) GetTransactionData() []byte {
	return mc.transactionData
}

func (mc *MockContext) GetFee() uint64 {
	return mc.Fee
}
